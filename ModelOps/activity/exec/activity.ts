/*
 * Copyright © 2020. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
import {Observable} from "rxjs/Observable";
import {Injectable, Injector, Inject} from "@angular/core";
import {Http} from "@angular/http";
import {
    WiContrib,
    WiServiceHandlerContribution,
    IValidationResult,
    ValidationResult,
    IFieldDefinition,
    IActivityContribution,
    IConnectorContribution,
    WiContributionUtils
} from "wi-studio/app/contrib/wi-contrib";

@WiContrib({})
@Injectable()
export class ExecActivityHandler extends WiServiceHandlerContribution {
	selectedConnector: string;
		
    constructor(@Inject(Injector) injector, private http: Http) {
        super(injector, http);
    }

    value = (fieldName: string, context: IActivityContribution): Observable<any> | any => {

		console.log('[ExecActivityHandler::value] Build field : ', fieldName);
        if (fieldName === "execConnection") {
            //Connector Type must match with the category defined in connector.json
            return Observable.create(observer => {
                let connectionRefs = [];                
                WiContributionUtils.getConnections(this.http, "ModelOps").subscribe((data: IConnectorContribution[]) => {
                    data.forEach(connection => {
						let connector: string;
						let outbound: boolean; 
                        for (let i = 0; i < connection.settings.length; i++) {
                        		if(connection.settings[i].name === "name") {
								connector = connection.settings[i].value
							} else if (connection.settings[i].name === "outbound") {
								outbound = connection.settings[i].value
							}
                        }
                        
                        //console.log("XXXXXXXXXXXXX 1 XXXXXXXXXXXXXXXXXX")
                        //console.log("connector -> " + connector)
                        //console.log("outbound -> " + outbound)
                        //console.log("XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX")
                        
                        
                        if(!outbound) {
                        	
                        //console.log("XXXXXXXXXXXXXXX 2 XXXXXXXXXXXXXXXX")
                        //console.log("connector -> " + connector)
                        //console.log("outbound -> " + outbound)
                        //console.log("XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX")
                        
                        		connectionRefs.push({
                        			"unique_id": WiContributionUtils.getUniqueId(connection),
                        			"name": connector
                        		});
                        }
                    });
                    observer.next(connectionRefs);
                });
            });
        }else if (fieldName === "Executable") {
        	var commandSchema = {
				"Executions" : 	{},
				"SystemEnvs" : {}
			}
			
			let numOfExecutions: IFieldDefinition = context.getField("numOfExecutions");
           	for (var i = 0; i < numOfExecutions.value; i++) {
				commandSchema["Executions"]["Execution_"+i] = populateAttribute("String")
			}
			
            let attrNames: IFieldDefinition = context.getField("SystemEnv");
            if (attrNames.value) {
                let data = JSON.parse(attrNames.value);
                for (var i = 0; i < data.length; i++) {
                		if ( data[i].PerCommand === "Yes") {
                			commandSchema["SystemEnvs"][data[i].Key] = populateAttribute("String")
					}
                }
            }
		console.log('[ExecActivityHandler::value] commandSchema : ', commandSchema);
            return JSON.stringify(commandSchema);
        }else if (fieldName === "Variables") {
        	var dataJsonSchema = {}
            let attrNames: IFieldDefinition = context.getField("variablesDef");
            if (attrNames.value) {
                let data = JSON.parse(attrNames.value);
                for (var i = 0; i < data.length; i++) {
                		dataJsonSchema[data[i].Name] = populateAttribute(data[i].Type);
                }
            }
            return JSON.stringify(dataJsonSchema);
        }else if (fieldName === "Result") {
        	var data = [{
				"Command" : populateAttribute("String"),
				"StdOut" : populateAttribute("String"),
				"StdErr" : populateAttribute("String")
			}]
            return JSON.stringify(data);
        }
		
        return null;
    }
 
    validate = (fieldName: string, context: IActivityContribution): Observable<IValidationResult> | IValidationResult => {
		console.log('[ExecActivityHandler::value] Validate field : ', fieldName);
		return null; 
    }
}

function populateAttribute(attrType) : any {
	switch(attrType) {
		case "Double" :
    			return 2.0;
		case "Integer":
			return 2;
		case "Long":
			return 2;
		case "Boolean":
			return true;
		case "Date":
			return 2;
		default:
    		return "2";
	}
}