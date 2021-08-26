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
export class ToDockerComposeActivityHandler extends WiServiceHandlerContribution {
	selectedConnector: string;
		
    constructor(@Inject(Injector) injector, private http: Http) {
        super(injector, http);
    }

    value = (fieldName: string, context: IActivityContribution): Observable<any> | any => {

		console.log('[ToDockerComposeActivityHandler::value] Build field : ', fieldName);
        if (fieldName === "DataFlow") {
            let dataflowObj = [{
	 			"Upstream": populateAttribute("String"),
				"Downstream" : populateAttribute("String")
			}];
		console.log('[ToDockerComposeActivityHandler::value] dataflowObj : ', dataflowObj);
            	return JSON.stringify(dataflowObj);
        } else if (fieldName === "Components") {
            let componentsObj = [{
	 			"Type": populateAttribute("String"),
				"Runtime" : populateAttribute("String"),
				"Name" : populateAttribute("String"),
				"Replicas" : populateAttribute("Integer"),
				"Volumes" : [{
					"Name" : populateAttribute("String"),
					"MountPoint" : populateAttribute("String")
				}],
            		Properties : [{
					Name: populateAttribute("String"),
					Value: populateAttribute("String")
				}]
			}];
		console.log('[ToDockerComposeActivityHandler::value] componentsObj : ', componentsObj);
            	return JSON.stringify(componentsObj);
        } else if (fieldName === "System")	{
            let systemObj = {
				Volume: [{
					Name: populateAttribute("String"),
					Value: populateAttribute("String")
				}]
			};
		console.log('[ToDockerComposeActivityHandler::value] systemObj : ', systemObj);
            	return JSON.stringify(systemObj);		
        }
		
        return null;
    }
 
    validate = (fieldName: string, context: IActivityContribution): Observable<IValidationResult> | IValidationResult => {
		console.log('[ToDockerComposeActivityHandler::value] Validate field : ', fieldName);
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