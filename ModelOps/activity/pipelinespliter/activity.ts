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
export class PipelineSpliterActivityHandler extends WiServiceHandlerContribution {
	selectedConnector: string;
		
    constructor(@Inject(Injector) injector, private http: Http) {
        super(injector, http);
    }

    value = (fieldName: string, context: IActivityContribution): Observable<any> | any => {
		console.log('[ModelOpsCMDConverter::value] Build field : ', fieldName);
		
        if (fieldName === "DataFlow")	{
            let dataflowObj = [{
				Upstream: populateAttribute("String"),
				Downstream: populateAttribute("String"),
			}];
			console.log('[PipelineSpliter::value] dataflowObj : ', dataflowObj);
            	return JSON.stringify(dataflowObj);		
        } else if (fieldName === "PipelineConfig")	{
            let configObj = [{
				Name: populateAttribute("String"),
				Type: populateAttribute("String"),
				ComponentConfig: populateAttribute("String"),
				Properties: [{
					Name: populateAttribute("String"),
					Value: populateAttribute("String")
				}]
			}];
			console.log('[PipelineSpliter::value] configObj : ', configObj);
            	return JSON.stringify(configObj);		
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