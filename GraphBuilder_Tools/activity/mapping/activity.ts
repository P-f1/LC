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
export class MappingContributionHandler extends WiServiceHandlerContribution {
	filename: string;
	content: string; 

    constructor(@Inject(Injector) injector, private http: Http) {
        super(injector, http);
    }

    value = (fieldName: string, context: IActivityContribution): Observable<any> | any => {
    	
		console.log('[MappingContributionHandler::value] fieldName = ' + fieldName);
        let attrNames: IFieldDefinition = context.getField("MappingFields");
		let isArray: IFieldDefinition = context.getField("IsArray")

		if (fieldName === "Mapping") {
            var attrJsonSchema = {};
            if (attrNames.value) {
                let data = JSON.parse(attrNames.value);
                for (var i = 0; i < data.length; i++) {
                		attrJsonSchema[data[i].Name] = populateAttribute(data[i].Type);
                }
				attrJsonSchema["SkipCondition"] = false;
            }
            
            if (isArray.value === true) {
				attrJsonSchema["ArraySize"] = 	2;
			}

			console.log('[MappingContributionHandler::value] attrJsonSchema = ' + JSON.stringify(attrJsonSchema));
            return JSON.stringify(attrJsonSchema);
        } else if (fieldName === "Data") {
        	if (isArray.value === true) {
            	var attrJsonSchemas = [];
            	if (attrNames.value) {
					var attrJsonSchema = {};
                	let data = JSON.parse(attrNames.value);
                	for (var i = 0; i < data.length; i++) {
                		attrJsonSchema[data[i].Name] = populateAttribute(data[i].Type);
                	}
					attrJsonSchemas.push(attrJsonSchema);
            	}
				console.log('[MappingContributionHandler::value] attrJsonSchemas = ' + JSON.stringify(attrJsonSchemas));
            	return JSON.stringify(attrJsonSchemas);
			} else {
            	var attrJsonSchema = {};
            	if (attrNames.value) {
                	let data = JSON.parse(attrNames.value);
                	for (var i = 0; i < data.length; i++) {
                		attrJsonSchema[data[i].Name] = populateAttribute(data[i].Type);
                	}
            	}
				console.log('[MappingContributionHandler::value] attrJsonSchema = ' + JSON.stringify(attrJsonSchema));
            	return JSON.stringify(attrJsonSchema);
			}
        }
        
        return null;
    }
 
    validate = (fieldName: string, context: IActivityContribution): Observable<IValidationResult> | IValidationResult => {
		
		console.log('[MappingContributionHandler::validate] fieldName = ' + fieldName);

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
		case "Object":
			return {};
		case "Array":
			return [];
		default:
    		return "2";
	}
}