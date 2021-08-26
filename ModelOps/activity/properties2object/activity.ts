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
export class Properties2ObjectActivityyHandler extends WiServiceHandlerContribution {
	selectedConnector: string;
		
    constructor(@Inject(Injector) injector, private http: Http) {
        super(injector, http);
    }

    value = (fieldName: string, context: IActivityContribution): Observable<any> | any => {

		console.log('[Properties2Object::value] Build field : ', fieldName);
		
		if (fieldName === "OverrideProperties") {
			let overridePropertiesObj = {};
			let attrNames: IFieldDefinition = context.getField("defOverrideProperties");
			if (attrNames.value) {
				let data = JSON.parse(attrNames.value);
				for (let i = 0; i < data.length; i++) {
					let splitted = data[i].Name.split(".");
					let current = overridePropertiesObj;
					for (let j = 0; j < splitted.length; j++) {
						console.log('[Properties2Object::value] splitted[j] = ', splitted[j]);
						if (splitted[j].endsWith("[]")) {
							console.log('[Properties2Object::value] an array ... ');
							let key = splitted[j].slice(0, splitted[j].indexOf("["));
							if (!(key in current)) {
								current[key] = [];
							}
							if (j==splitted.length-1) {
								/* It's an primitive array element.*/
                				current[key].push(populateAttribute(data[i].Type));
							} else {
								if (current[key].length<1) {
									current[key].push({})
								}
								current = current[key][0];
							}
						} else {
							console.log('[Properties2Object::value] an object ... ');
							if (j==splitted.length-1) {
								/* It's an object attribute.*/
                				current[splitted[j]] = populateAttribute(data[i].Type);
							} else {
								if (!(splitted[j] in current)) {
									current[splitted[j]] = {};
								}
								current = current[splitted[j]];
							}
						}
					}
				}
			}
            return JSON.stringify(overridePropertiesObj);
        } else if (fieldName === "Properties") {
            let propertiesObj = [{
	 			"Type": populateAttribute("String"),
				"Name" : populateAttribute("String"),
				"Value" : populateAttribute("String")
			}];
			return JSON.stringify(propertiesObj);
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
        } else if (fieldName === "PassThroughData") {
            var attrJsonSchema = {};
            let attrNames: IFieldDefinition = context.getField("PassThrough");
            if (attrNames.value) {
                let data = JSON.parse(attrNames.value);
                for (var i = 0; i < data.length; i++) {
                	    attrJsonSchema[data[i].FieldName] = populateAttribute(data[i].Type);
                }
            }
            return JSON.stringify(attrJsonSchema);
        } else if (fieldName === "PassThroughDataOut") {
            var attrJsonSchema = {};
            let attrNames: IFieldDefinition = context.getField("PassThrough");
            if (attrNames.value) {
                let data = JSON.parse(attrNames.value);
                for (var i = 0; i < data.length; i++) {
                	    attrJsonSchema[data[i].FieldName] = populateAttribute(data[i].Type);
                }
            }
            return JSON.stringify(attrJsonSchema);
        }
		
        return null;
    }
 
    validate = (fieldName: string, context: IActivityContribution): Observable<IValidationResult> | IValidationResult => {
		console.log('[Properties2Object::value] Validate field : ', fieldName);
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