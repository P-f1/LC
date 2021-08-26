"use strict";
var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
Object.defineProperty(exports, "__esModule", { value: true });
var core_1 = require("@angular/core");
var http_1 = require("@angular/http");
var Observable_1 = require("rxjs/Observable");
var protobufjs = require("protobufjs");
var wi_contrib_1 = require("wi-studio/app/contrib/wi-contrib");
var contrib_1 = require("wi-studio/common/models/contrib");
var lodash = require("lodash");
var emptyArray = [];
var protoMap = new Map();
var oldProtoFileContent;
var serviceNames = [];
var camelCaseRe = /_([a-z])/g;
var grpcHandler = (function (_super) {
    __extends(grpcHandler, _super);
    function grpcHandler(injector, http, contribModelService) {
        var _this = _super.call(this, injector, http, contribModelService) || this;
        _this.injector = injector;
        _this.http = http;
        _this.contribModelService = contribModelService;
        _this.value = function (fieldName, context) {
            var protoFile = context.getField("protoFile").value;
            var serviceName = context.getField("serviceName").value;
            var methodName = context.getField("methodName").value;
            switch (fieldName) {
                case "protoName":
                    if (protoFile) {
                        return protoFile.filename;
                    }
                case "serviceName":
                    if (protoFile) {
                        var protoFileContent = protoFile.content;
                        if (oldProtoFileContent != protoFileContent) {
                            oldProtoFileContent = protoFileContent;
                            if (protoMap.has(protoFileContent)) {
                                var grpcResult = protoMap.get(protoFileContent);
                                serviceNames = grpcResult.success ? Object.keys(grpcResult.services) : emptyArray;
                            }
                            else {
                                var grpcResult = _this.parseProtoFile(protoFile);
                                serviceNames = grpcResult.success ? Object.keys(grpcResult.services) : emptyArray;
                            }
                        }
                        return serviceNames;
                    }
                    return emptyArray;
                case "methodName":
                    if (protoFile && serviceName && serviceName != "") {
                        var grpcResult = protoMap.get(protoFile.content);
                        return Object.keys(grpcResult.services[serviceName].methods);
                    }
                    return emptyArray;
                case "params":
                    if (protoFile && serviceName && methodName && methodName != "") {
                        var grpcResult = protoMap.get(protoFile.content);
                        return grpcResult.services[serviceName].methods[methodName].inputs;
                    }
                    return null;
                case "data":
                    if (protoFile && serviceName && methodName && methodName != "") {
                        var grpcResult = protoMap.get(protoFile.content);
                        return grpcResult.services[serviceName].methods[methodName].outputs;
                    }
                    return null;
                case "enableTLS":
                    return Observable_1.Observable.create(function (observer) {
                        wi_contrib_1.WiContributionUtils.getAppConfig(_this.http)
                            .subscribe(function (data) {
                            if (data.deployment === wi_contrib_1.APP_DEPLOYMENT.CLOUD) {
                                observer.next(false);
                            }
                            else {
                                observer.next(null);
                            }
                        }, function (err) {
                            observer.next(wi_contrib_1.ValidationResult.newValidationResult().setVisible(false));
                        }, function () { return observer.complete(); });
                    });
                default:
                    return null;
            }
        };
        _this.validate = function (fieldName, context) {
            if (fieldName === "port" || fieldName === "enableTLS") {
                return Observable_1.Observable.create(function (observer) {
                    wi_contrib_1.WiContributionUtils.getAppConfig(_this.http)
                        .subscribe(function (data) {
                        if (data.deployment === wi_contrib_1.APP_DEPLOYMENT.ON_PREMISE) {
                            observer.next(wi_contrib_1.ValidationResult.newValidationResult().setVisible(true));
                        }
                        else {
                            observer.next(wi_contrib_1.ValidationResult.newValidationResult().setVisible(false));
                        }
                    }, function (err) {
                        observer.next(wi_contrib_1.ValidationResult.newValidationResult().setVisible(false));
                    }, function () { return observer.complete(); });
                });
            }
            else if (fieldName === "protoFile") {
                var protoFile = context.getField("protoFile").value;
                if (protoFile) {
                    if (protoMap && protoMap.has(protoFile.content)) {
                        var grpcResult = protoMap.get(protoFile.content);
                        if (!grpcResult.success) {
                            return wi_contrib_1.ValidationResult.newValidationResult().setError("gRPCError", grpcResult.error);
                        }
                    }
                }
                else {
                    return wi_contrib_1.ValidationResult.newValidationResult().setError("Required", "A proto file must be configured.");
                }
            }
            else if (fieldName === "protoName") {
                return wi_contrib_1.ValidationResult.newValidationResult().setVisible(false);
            }
            else if (fieldName === "serverCert") {
                var secure = context.getField("enableTLS").value;
                if (secure) {
                    var valResult = wi_contrib_1.ValidationResult.newValidationResult();
                    valResult.setVisible(true);
                    var serverCert = context.getField("serverCert").value;
                    if (!serverCert) {
                        valResult.setError("Required", "A CA or Server Certificate file must be configured.");
                    }
                    return valResult;
                }
                else {
                    return wi_contrib_1.ValidationResult.newValidationResult().setVisible(false);
                }
            }
            else if (fieldName === "serverKey") {
                var secure = context.getField("enableTLS").value;
                if (secure) {
                    var valResult = wi_contrib_1.ValidationResult.newValidationResult();
                    valResult.setVisible(true);
                    var serverKey = context.getField("serverKey").value;
                    if (!serverKey) {
                        valResult.setError("Required", "A Server Key file must be configured.");
                    }
                    return valResult;
                }
                else {
                    return wi_contrib_1.ValidationResult.newValidationResult().setVisible(false);
                }
            }
            else if (fieldName === "content" || fieldName === "grpcData" || fieldName === "code") {
                return wi_contrib_1.ValidationResult.newValidationResult().setVisible(false);
            }
            return null;
        };
        _this.action = function (actionId, context) {
            var modelService = _this.getModelService();
            var result = wi_contrib_1.CreateFlowActionResult.newActionResult();
            var actionResult = wi_contrib_1.ActionResult.newActionResult();
            var protoFile = context.getField("protoFile").value;
            return Observable_1.Observable.create(function (observer) {
                var grpcResult = _this.parseProtoFile(protoFile);
                if (!grpcResult.success) {
                    actionResult.setSuccess(false);
                    actionResult.setResult(wi_contrib_1.ValidationResult.newValidationResult().setError("gRPCError", grpcResult.error));
                }
                else {
                    if (context.getMode() === contrib_1.MODE.SERVERLESS_FLOW) {
                        var _serviceName = context.getField("serviceName").value;
                        var _methodName = context.getField("methodName").value;
                        var trigger = _this.doTriggerConfiguration(context, grpcResult, _serviceName, _methodName);
                        var dummyFlowModel = modelService.createFlow(context.getFlowName(), context.getFlowDescription(), false);
                        result = result.addTriggerFlowMapping(lodash.cloneDeep(trigger), lodash.cloneDeep(dummyFlowModel));
                    }
                    else if (context.getMode() === contrib_1.MODE.UPLOAD) {
                        Object.keys(grpcResult.services).map(function (s) {
                            Object.keys(grpcResult.services[s].methods).map(function (m) {
                                var trigger = _this.doTriggerConfiguration(context, grpcResult, s, m);
                                var flowModel = modelService.createFlow(s + "_" + m, "", false);
                                var returnAct = modelService.createFlowElement("Default/flogo-return");
                                var uploadFlow = flowModel.addFlowElement(returnAct);
                                result = result.addTriggerFlowMapping(lodash.cloneDeep(trigger), lodash.cloneDeep(uploadFlow));
                            });
                        });
                    }
                    actionResult.setSuccess(true).setResult(result);
                }
                observer.next(actionResult);
            });
        };
        return _this;
    }
    grpcHandler.prototype.doTriggerConfiguration = function (context, grpcResult, serviceName, methodName) {
        var _port = context.getField("port").value;
        var _protoName = context.getField("protoName").value;
        var _enableTLS = context.getField("enableTLS").value;
        var _serverCert = context.getField("serverCert").value;
        var _serverKey = context.getField("serverKey").value;
        var _protoFile = context.getField("protoFile").value;
        var trigger = this.getModelService().createTriggerElement("Default/grpc-trigger");
        if (trigger && trigger.settings && trigger.settings.length > 0) {
            trigger.settings.map(function (setting) {
                if (setting.name === "port") {
                    setting.value = _port;
                }
                else if (setting.name === "protoName") {
                    setting.value = _protoName;
                }
                else if (setting.name === "enableTLS") {
                    setting.value = _enableTLS;
                }
                else if (setting.name === "serverCert") {
                    setting.value = _serverCert;
                }
                else if (setting.name === "serverKey") {
                    setting.value = _serverKey;
                }
                else if (setting.name === "protoFile") {
                    setting.value = _protoFile;
                }
            });
        }
        if (trigger && trigger.handler && trigger.handler.settings && trigger.handler.settings.length > 0) {
            trigger.handler.settings.map(function (hndlrSetting) {
                if (hndlrSetting.name === "serviceName") {
                    hndlrSetting.value = serviceName;
                }
                else if (hndlrSetting.name === "methodName") {
                    hndlrSetting.value = methodName;
                }
            });
        }
        if (trigger && trigger.outputs && trigger.outputs.length > 0) {
            for (var j = 0; j < trigger.outputs.length; j++) {
                if (trigger.outputs[j].name === "params") {
                    trigger.outputs[j].value = grpcResult.services[serviceName].methods[methodName].inputs;
                    break;
                }
            }
        }
        if (trigger && trigger.reply && trigger.reply.length > 0) {
            for (var j = 0; j < trigger.reply.length; j++) {
                if (trigger.reply[j].name === "data") {
                    trigger.reply[j].value = grpcResult.services[serviceName].methods[methodName].outputs;
                    break;
                }
            }
        }
        this.doTriggerMapping(trigger);
        return trigger;
    };
    grpcHandler.prototype.doTriggerMapping = function (trigger) {
        var outputMappingElement = this.contribModelService.createMapping();
        var outputexpr = this.contribModelService.createMapExpression();
        for (var j = 0; j < trigger.outputs.length; j++) {
            if (trigger.outputs[j].name === "params") {
                outputMappingElement.addMapping("$INPUT['" + trigger.outputs[j].name + "']", outputexpr.setExpression("$trigger." + trigger.outputs[j].name));
                break;
            }
        }
        trigger.inputMappings = outputMappingElement;
        var replyMappingElement = this.contribModelService.createMapping();
        var replyexpr = this.contribModelService.createMapExpression();
        for (var i = 0; i < trigger.reply.length; i++) {
            if (trigger.reply[i].name === "data") {
                replyMappingElement.addMapping("$INPUT['" + trigger.reply[i].name + "']", replyexpr.setExpression("$flow." + trigger.reply[i].name));
                break;
            }
        }
        trigger.outputMappings = replyMappingElement;
    };
    grpcHandler.prototype.parseServiceNames = function (root, packageName) {
        var tempJson = root.toJSON();
        var package_array = packageName.split(".");
        var nestedKeys = [];
        var i = 0;
        while (true) {
            nestedKeys = Object.keys(tempJson.nested);
            if (nestedKeys.length > 1) {
                break;
            }
            tempJson = tempJson.nested[package_array[i++]];
        }
        var serviceNames = [];
        nestedKeys.forEach(function (nk) {
            var nestedObj = tempJson.nested[nk];
            if (nestedObj["methods"]) {
                serviceNames.push(nk);
            }
        });
        return serviceNames;
    };
    grpcHandler.prototype.parseProtoFile = function (protoFile) {
        var _this = this;
        var result = {};
        var protoContentEncoded = protoFile.content.split(",")[1];
        var protoContent = atob(protoContentEncoded);
        try {
            var parsedResult = protobufjs.parse(protoContent, { keepCase: true });
            if (parsedResult && parsedResult.syntax != "proto2") {
                result.success = true;
                result.services = {};
                try {
                    var packageName = "";
                    if (parsedResult.package) {
                        packageName = parsedResult.package + ".";
                    }
                    var serviceNames_1 = this.parseServiceNames(parsedResult.root, packageName);
                    if (serviceNames_1 && serviceNames_1.length > 0) {
                        var _loop_1 = function (i) {
                            var grpcService = {};
                            grpcService.methods = {};
                            var service = parsedResult.root.lookupService(packageName + serviceNames_1[i]);
                            service.methodsArray.forEach(function (m) {
                                var grpcMethod = {};
                                grpcMethod.name = _this.convertToCamelCase(m.name);
                                grpcMethod.inputs = _this.parseProtoMethod(m, "input");
                                grpcMethod.outputs = _this.parseProtoMethod(m, "output");
                                grpcService.methods[grpcMethod.name] = grpcMethod;
                            });
                            result.services[this_1.convertToCamelCase(service.name)] = grpcService;
                        };
                        var this_1 = this;
                        for (var i = 0; i < serviceNames_1.length; i++) {
                            _loop_1(i);
                        }
                    }
                    else {
                        result.success = false;
                        result.error = "Error: Could not find any service definition in proto file.";
                    }
                }
                catch (e) {
                    result.success = false;
                    result.error = e + "";
                }
            }
            else {
                result.success = false;
                result.error = "Error: proto2 syntax is not supported. Please define the proto file using proto3 syntax.";
            }
        }
        catch (e) {
            result.success = false;
            result.error = e + "";
        }
        protoMap.set(protoFile.content, result);
        return result;
    };
    grpcHandler.prototype.parseProtoMethod = function (m, paramType) {
        var _this = this;
        var schema = {
            "type": "object",
            "properties": {},
            "required": []
        };
        if (!m.resolved)
            m.resolve();
        var methodFields;
        if (paramType === "input") {
            methodFields = m.resolvedRequestType.fieldsArray;
        }
        else {
            methodFields = m.resolvedResponseType.fieldsArray;
        }
        methodFields.forEach(function (f) {
            if (!f.resolved)
                f.resolve();
            if (f.rule === "repeated") {
                schema.properties[f.name] = { type: "array" };
                schema.properties[f.name].items = _this.coerceField(f);
            }
            else {
                schema.properties[f.name] = _this.coerceField(f);
            }
        });
        return JSON.stringify(schema);
    };
    grpcHandler.prototype.coerceField = function (f) {
        if (this.isScalarType(f.type)) {
            return this.coerceType(f.type);
        }
        else {
            return this.coerceType(f.resolvedType);
        }
    };
    grpcHandler.prototype.isScalarType = function (_type) {
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
    };
    grpcHandler.prototype.coerceType = function (baseType) {
        var _this = this;
        if (baseType instanceof protobufjs.Enum) {
            if (!baseType.resolved)
                baseType.resolve();
            return { type: "string", enum: Object.keys(baseType.values) };
        }
        else if (baseType instanceof protobufjs.Type) {
            if (!baseType.resolved)
                baseType.resolve();
            var schema_1 = {
                "type": "object",
                "properties": {}
            };
            baseType.fieldsArray.forEach(function (f) {
                if (!f.resolved)
                    f.resolve();
                if (f.rule === "repeated") {
                    schema_1.properties[f.name] = { type: "array" };
                    schema_1.properties[f.name].items = _this.coerceField(f);
                }
                else {
                    schema_1.properties[f.name] = _this.coerceField(f);
                }
            });
            return schema_1;
        }
        else {
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
    };
    grpcHandler.prototype.convertToCamelCase = function (str) {
        return str.substring(0, 1).toUpperCase() + str.substring(1).replace(camelCaseRe, function ($0, $1) { return $1.toUpperCase(); });
    };
    grpcHandler = __decorate([
        wi_contrib_1.WiContrib({}),
        core_1.Injectable(),
        __metadata("design:paramtypes", [core_1.Injector, http_1.Http, wi_contrib_1.WiContribModelService])
    ], grpcHandler);
    return grpcHandler;
}(wi_contrib_1.WiServiceHandlerContribution));
exports.grpcHandler = grpcHandler;
//# sourceMappingURL=grpcHandler.js.map