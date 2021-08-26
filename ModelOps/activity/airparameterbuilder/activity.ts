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
export class AirParameterBuilderActivityHandler extends WiServiceHandlerContribution {
	selectedConnector: string;
		
    constructor(@Inject(Injector) injector, private http: Http) {
        super(injector, http);
    }

    value = (fieldName: string, context: IActivityContribution): Observable<any> | any => {
		console.log('[AirParameterBuilder::value] Build field : ', fieldName);		
        if (fieldName === "Descriptor") {
            let descriptor = {
            	"FlogoDescriptor" : populateAttribute("String"),
				"F1Properties" : [{
					"Group" : populateAttribute("String"),
					"Value" : {
	 					"Name": populateAttribute("String"),
						"Value" : populateAttribute("String"),
						"Type" : populateAttribute("String")
					}
				}]
            };
            return JSON.stringify(descriptor);
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
        }
        return null;
    }
 
    validate = (fieldName: string, context: IActivityContribution): Observable<IValidationResult> | IValidationResult => {
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