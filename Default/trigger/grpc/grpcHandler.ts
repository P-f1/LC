import { Injectable, Injector } from "@angular/core";
import { Http } from "@angular/http";
import { Observable } from "rxjs/Observable";
import * as protobufjs from "protobufjs";
import {
    WiContrib,
    WiServiceHandlerContribution,
    IValidationResult,
    ValidationResult,
    ActionResult,
    IActionResult,
    WiContribModelService,
    WiContributionUtils,
    CreateFlowActionResult,
    APP_DEPLOYMENT
} from "wi-studio/app/contrib/wi-contrib";
import { ITriggerContribution, IFieldDefinition, MODE, ICreateFlowActionContext, ITriggerElement } from "wi-studio/common/models/contrib";
import * as lodash from "lodash";


const emptyArray = [];
let protoMap = new Map(); // Saves the IGrpcResult for a specific encoded protoContent
let oldProtoFileContent: string;
let serviceNames: string[] = [];
let camelCaseRe = /_([a-z])/g;

@WiContrib({})
@Injectable()
export class grpcHandler extends WiServiceHandlerContribution {

    constructor(private injector: Injector, private http: Http, private contribModelService: WiContribModelService) {
        super(injector, http, contribModelService);
    }

    value = (fieldName: string, context: ITriggerContribution): Observable<any> | any => {
        let protoFile = context.getField("protoFile").value;
        let serviceName: string = context.getField("serviceName").value;
        let methodName: string = context.getField("methodName").value;
        switch (fieldName) {
            case "protoName":
                if (protoFile) {
                    return protoFile.filename;
                }
            case "serviceName":
                if (protoFile) {
                    // parse protoFile to find service names
                    let protoFileContent: string = protoFile.content;
                    if (oldProtoFileContent != protoFileContent) {
                        oldProtoFileContent = protoFileContent;
                        // implies change in proto file
                        if (protoMap.has(protoFileContent)) {
                            // if parse was successful, return service names
                            let grpcResult: IGrpcResult = protoMap.get(protoFileContent);
                            serviceNames = grpcResult.success ? Object.keys(grpcResult.services) : emptyArray;
                        } else {
                            // new protofile, parse proto content
                            let grpcResult = this.parseProtoFile(protoFile);
                            serviceNames = grpcResult.success ? Object.keys(grpcResult.services) : emptyArray;
                        }
                    }
                    return serviceNames;
                }
                return emptyArray;
            case "methodName":
                if (protoFile && serviceName && serviceName != "") {
                    // if serviceName is not empty implies parsing was successfull earlier
                    let grpcResult: IGrpcResult = protoMap.get(protoFile.content);
                    return Object.keys(grpcResult.services[serviceName].methods);
                }
                return emptyArray;
            case "params":
                if (protoFile && serviceName && methodName && methodName != "") {
                    // find the methodName for the service and return the inputs
                    let grpcResult: IGrpcResult = protoMap.get(protoFile.content);
                    return grpcResult.services[serviceName].methods[methodName].inputs;
                }
                return null;
            case "data":
                if (protoFile && serviceName && methodName && methodName != "") {
                    // find the methodName for the service and return the outputs
                    let grpcResult: IGrpcResult = protoMap.get(protoFile.content);
                    return grpcResult.services[serviceName].methods[methodName].outputs;
                }
                return null;
            case "enableTLS":
                // change value to false (in case of imports from FE to TCI)
                return Observable.create(observer => {
                    WiContributionUtils.getAppConfig(this.http)
                        .subscribe(data => {
                            if (data.deployment === APP_DEPLOYMENT.CLOUD) {
                                observer.next(false);
                            } else {
                                observer.next(null);
                            }
                        }, err => {
                            observer.next(ValidationResult.newValidationResult().setVisible(false));
                        }, () => observer.complete());
                });
            default:
                return null;
        }
    }

    validate = (fieldName: string, context: ITriggerContribution): Observable<IValidationResult> | IValidationResult => {
        if (fieldName === "port" || fieldName === "enableTLS") {
            return Observable.create(observer => {
                WiContributionUtils.getAppConfig(this.http)
                    .subscribe(data => {
                        if (data.deployment === APP_DEPLOYMENT.ON_PREMISE) {
                            // port and enableTLS visible only on FE
                            observer.next(ValidationResult.newValidationResult().setVisible(true));
                        } else {
                            observer.next(ValidationResult.newValidationResult().setVisible(false));
                        }
                    }, err => {
                        observer.next(ValidationResult.newValidationResult().setVisible(false));
                    }, () => observer.complete());
            });
        } else if (fieldName === "protoFile") {
            let protoFile = context.getField("protoFile").value;
            if (protoFile) {
                // check if proto file is valid
                if (protoMap && protoMap.has(protoFile.content)) {
                    let grpcResult: IGrpcResult = protoMap.get(protoFile.content);
                    if (!grpcResult.success) {
                        return ValidationResult.newValidationResult().setError("gRPCError", grpcResult.error);
                    }
                }
            } else {
                return ValidationResult.newValidationResult().setError("Required", "A proto file must be configured.");
            }
        } else if (fieldName === "protoName") {
            // if FE or TCI -- hide field and fill value for json
            return ValidationResult.newValidationResult().setVisible(false);
        } else if (fieldName === "serverCert") {
            let secure: boolean = context.getField("enableTLS").value;
            if (secure) {
                let valResult = ValidationResult.newValidationResult();
                valResult.setVisible(true);
                let serverCert = context.getField("serverCert").value;
                if (!serverCert) {
                    valResult.setError("Required", "A CA or Server Certificate file must be configured.")
                }
                return valResult;
            } else {
                return ValidationResult.newValidationResult().setVisible(false);
            }
        } else if (fieldName === "serverKey") {
            let secure: boolean = context.getField("enableTLS").value;
            if (secure) {
                let valResult = ValidationResult.newValidationResult();
                valResult.setVisible(true);
                let serverKey = context.getField("serverKey").value;
                if (!serverKey) {
                    valResult.setError("Required", "A Server Key file must be configured.")
                }
                return valResult;
            } else {
                return ValidationResult.newValidationResult().setVisible(false);
            }
        } else if (fieldName === "content" || fieldName === "grpcData" || fieldName === "code") {
            return ValidationResult.newValidationResult().setVisible(false);
        }
        return null;
    }

    action = (actionId: string, context: ICreateFlowActionContext): Observable<IActionResult> | IActionResult => {
        let modelService = this.getModelService();
        let result = CreateFlowActionResult.newActionResult();
        let actionResult = ActionResult.newActionResult();
        let protoFile: IFieldDefinition = context.getField("protoFile").value;
        return Observable.create(observer => {
            // Parse protofile
            let grpcResult = this.parseProtoFile(protoFile);
            if (!grpcResult.success) {
                // validation error
                actionResult.setSuccess(false);
                actionResult.setResult(ValidationResult.newValidationResult().setError("gRPCError", grpcResult.error));
            } else {
                if (context.getMode() === MODE.SERVERLESS_FLOW) {
                    let _serviceName: string = context.getField("serviceName").value;
                    let _methodName: string = context.getField("methodName").value;

                    // create trigger
                    let trigger = this.doTriggerConfiguration(context, grpcResult, _serviceName, _methodName);

                    let dummyFlowModel = modelService.createFlow(context.getFlowName(), context.getFlowDescription(), false);
                    result = result.addTriggerFlowMapping(lodash.cloneDeep(trigger), lodash.cloneDeep(dummyFlowModel));
                } else if (context.getMode() === MODE.UPLOAD) {
                    // create a flow and trigger for each method in each service of proto file
                    Object.keys(grpcResult.services).map(s => {
                        Object.keys(grpcResult.services[s].methods).map(m => {
                            // create trigger
                            let trigger = this.doTriggerConfiguration(context, grpcResult, s, m);

                            // create flow
                            let flowModel = modelService.createFlow(s + "_" + m, "", false);

                            // add return activity
                            let returnAct = modelService.createFlowElement("Default/flogo-return");
                            let uploadFlow = flowModel.addFlowElement(returnAct);
                            result = result.addTriggerFlowMapping(lodash.cloneDeep(trigger), lodash.cloneDeep(uploadFlow));

                        });
                    });
                }
                actionResult.setSuccess(true).setResult(result);
            }
            observer.next(actionResult);
        });
    }

    doTriggerConfiguration(context: ICreateFlowActionContext, grpcResult: IGrpcResult, serviceName: string, methodName: string): ITriggerElement {
        // get values from trigger context
        let _port = context.getField("port").value;
        let _protoName = context.getField("protoName").value;
        let _enableTLS = context.getField("enableTLS").value;
        let _serverCert = context.getField("serverCert").value;
        let _serverKey = context.getField("serverKey").value;
        let _protoFile = context.getField("protoFile").value;

        let trigger = this.getModelService().createTriggerElement("Default/grpc-trigger");

        // copy values from trigger wizard into trigger
        if (trigger && trigger.settings && trigger.settings.length > 0) {
            // set all trigger settings -- use default values if no value provided
            trigger.settings.map(setting => {
                if (setting.name === "port") {
                    setting.value = _port;
                } else if (setting.name === "protoName") {
                    setting.value = _protoName;
                } else if (setting.name === "enableTLS") {
                    setting.value = _enableTLS;
                } else if (setting.name === "serverCert") {
                    setting.value = _serverCert;
                } else if (setting.name === "serverKey") {
                    setting.value = _serverKey;
                } else if (setting.name === "protoFile") {
                    setting.value = _protoFile;
                }
            });
        }

        if (trigger && trigger.handler && trigger.handler.settings && trigger.handler.settings.length > 0) {
            // set serviceName and methodName on handler
            trigger.handler.settings.map(hndlrSetting => {
                if (hndlrSetting.name === "serviceName") {
                    hndlrSetting.value = serviceName;
                } else if (hndlrSetting.name === "methodName") {
                    hndlrSetting.value = methodName;
                }
            });
        }

        // copy flow input schema to trigger output
        if (trigger && trigger.outputs && trigger.outputs.length > 0) {
            // set trigger output on params
            for (let j = 0; j < trigger.outputs.length; j++) {
                if (trigger.outputs[j].name === "params") {
                    // find the grpc method that matches methodName for the trigger
                    trigger.outputs[j].value = grpcResult.services[serviceName].methods[methodName].inputs;
                    break;
                }
            }
        }

        // copy flow output schema to trigger reply
        if (trigger && trigger.reply && trigger.reply.length > 0) {
            // set trigger reply
            for (let j = 0; j < trigger.reply.length; j++) {
                if (trigger.reply[j].name === "data") {
                    trigger.reply[j].value = grpcResult.services[serviceName].methods[methodName].outputs;
                    break;
                }
            }
        }

        // map trigger output to flow input and flow output to trigger reply for each flow
        this.doTriggerMapping(trigger);

        return trigger;
    }

    doTriggerMapping(trigger: ITriggerElement) {
        let outputMappingElement = this.contribModelService.createMapping();
        let outputexpr = this.contribModelService.createMapExpression();
        for (let j = 0; j < trigger.outputs.length; j++) {
            if (trigger.outputs[j].name === "params") {
                // map params to params
                outputMappingElement.addMapping("$INPUT['" + trigger.outputs[j].name + "']", outputexpr.setExpression("$trigger." + trigger.outputs[j].name));
                break;
            }

        }
        (<any>trigger).inputMappings = outputMappingElement;

        let replyMappingElement = this.contribModelService.createMapping();
        let replyexpr = this.contribModelService.createMapExpression();
        for (let i = 0; i < trigger.reply.length; i++) {
            if (trigger.reply[i].name === "data") {
                // map only data element
                replyMappingElement.addMapping("$INPUT['" + trigger.reply[i].name + "']", replyexpr.setExpression("$flow." + trigger.reply[i].name));
                break;
            }
        }
        (<any>trigger).outputMappings = replyMappingElement;
    }

    // parse servicenames from protofile
    parseServiceNames(root: protobufjs.Root, packageName: string): string[] {
        let tempJson = root.toJSON();
        let package_array = packageName.split(".");
        let nestedKeys = [];
        let i = 0;
        while (true) {
            nestedKeys = Object.keys(tempJson.nested);
            if (nestedKeys.length > 1) {
                // found service level
                break;
            }
            tempJson = tempJson.nested[package_array[i++]];
        }
        let serviceNames: string[] = [];
        nestedKeys.forEach(nk => {
            let nestedObj = tempJson.nested[nk];
            if (nestedObj["methods"]) {
                // implies object is a service
                serviceNames.push(nk);
            }
        });
        return serviceNames;
    }

    // parses proto file content and builds IGrpcResult
    parseProtoFile(protoFile: any): IGrpcResult {
        let result: IGrpcResult = <IGrpcResult>{};
        let protoContentEncoded: string = protoFile.content.split(",")[1];
        let protoContent = atob(protoContentEncoded);
        try {
            let parsedResult = protobufjs.parse(protoContent, { keepCase: true });
            if (parsedResult && parsedResult.syntax != "proto2") {
                result.success = true;
                result.services = <IGrpcServicesMap>{};
                // find service file
                try {
                    let packageName = "";
                    if (parsedResult.package) {
                        packageName = parsedResult.package + ".";
                    }
                    let serviceNames = this.parseServiceNames(parsedResult.root, packageName);
                    if (serviceNames && serviceNames.length > 0) {
                        for (let i = 0; i < serviceNames.length; i++) {
                            let grpcService: IGrpcService = <IGrpcService>{};
                            grpcService.methods = <IGrpcMethodsMap>{};
                            let service = parsedResult.root.lookupService(packageName + serviceNames[i]);
                            service.methodsArray.forEach(m => {
                                // build input and output schema for each method in service
                                let grpcMethod: IGrpcMethod = <IGrpcMethod>{};
                                // convert only methodName and servicename from camel_case to CamelCase to support go generated server files
                                grpcMethod.name = this.convertToCamelCase(m.name);
                                grpcMethod.inputs = this.parseProtoMethod(m, "input");
                                grpcMethod.outputs = this.parseProtoMethod(m, "output");
                                grpcService.methods[grpcMethod.name] = grpcMethod;
                            });
                            result.services[this.convertToCamelCase(service.name)] = grpcService;
                        }
                    } else {
                        // could not find service names so set error
                        result.success = false;
                        result.error = "Error: Could not find any service definition in proto file.";
                    }
                } catch (e) {
                    // Error finding service name
                    result.success = false;
                    result.error = e + "";
                }
            } else {
                result.success = false;
                result.error = "Error: proto2 syntax is not supported. Please define the proto file using proto3 syntax.";
            }
        } catch (e) {
            // Error parsing proto file
            result.success = false;
            result.error = e + "";
        }
        protoMap.set(protoFile.content, result);
        return result;
    }

    parseProtoMethod(m: protobufjs.Method, paramType: string): string {
        let schema = {
            "type": "object",
            "properties": {},
            "required": []
        };
        if (!m.resolved) m.resolve();
        let methodFields: protobufjs.Field[];
        if (paramType === "input") {
            methodFields = m.resolvedRequestType.fieldsArray;
        } else {
            methodFields = m.resolvedResponseType.fieldsArray;
        }
        methodFields.forEach(f => {
            if (!f.resolved) f.resolve();
            if (f.rule === "repeated") {
                // create array schema
                schema.properties[f.name] = { type: "array" };
                schema.properties[f.name].items = this.coerceField(f);
            } else {
                schema.properties[f.name] = this.coerceField(f);
            }
        });
        return JSON.stringify(schema);
    }

    coerceField(f: protobufjs.Field) {
        if (this.isScalarType(f.type)) {
            return this.coerceType(f.type);
        } else {
            return this.coerceType(f.resolvedType);
        }
    }

    isScalarType(_type: string) {
        switch (_type) {
            case "uint32":
            case "int32":
            case "sint32":
            case "int64":
            case "uint64":
            case "sint64":
            case "fixed32":
            case "sfixed32":
            case "fixed64":
            case "sfixed64":
            case "float":
            case "double":
            case "bool":
            case "bytes":
            case "string":
                return true;
            default: return false;
        }
    }

    coerceType(baseType: protobufjs.Type | protobufjs.Enum | string) {
        if (baseType instanceof protobufjs.Enum) {
            // return enum
            if (!baseType.resolved) baseType.resolve();
            return { type: "string", enum: Object.keys(baseType.values) };
        } else if (baseType instanceof protobufjs.Type) {
            // object type
            if (!baseType.resolved) baseType.resolve();

            let schema = {
                "type": "object",
                "properties": {}
            };

            baseType.fieldsArray.forEach(f => {
                if (!f.resolved) f.resolve();
                if (f.rule === "repeated") {
                    // create array schema
                    schema.properties[f.name] = { type: "array" };
                    schema.properties[f.name].items = this.coerceField(f);
                } else {
                    schema.properties[f.name] = this.coerceField(f);
                }

            });
            return schema;

        } else {
            // scalar type
            switch (baseType) {
                case "uint32":
                case "int32":
                case "sint32":
                case "int64":
                case "uint64":
                case "sint64":
                case "fixed32":
                case "sfixed32":
                case "fixed64":
                case "sfixed64":
                case "float":
                case "double":
                    return { type: "number" };
                case "bool":
                    return { type: "boolean" };
                case "bytes":
                case "string":
                    return { type: "string" };
                default:
                    return { type: "any" };
            }
        }
    }

    // converts camel_case to CamelCase
    convertToCamelCase(str: string) {
        return str.substring(0, 1).toUpperCase() + str.substring(1).replace(camelCaseRe, function ($0, $1) { return $1.toUpperCase(); });
    }
}
export interface IGrpcResult {
    success?: boolean;
    error?: string;
    services?: IGrpcServicesMap;
}

export interface IGrpcServicesMap {
    [service: string]: IGrpcService
}

export interface IGrpcService {
    methods?: IGrpcMethodsMap
}

export interface IGrpcMethodsMap {
    [method: string]: IGrpcMethod;
}

export interface IGrpcMethod {
    name: string;
    inputs?: any;
    outputs?: any;
}