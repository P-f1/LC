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
export class PipecouplerActivityContributionHandler extends WiServiceHandlerContribution {
	selectedConnector: string;
		
    constructor(@Inject(Injector) injector, private http: Http) {
        super(injector, http);
    }

    value = (fieldName: string, context: IActivityContribution): Observable<any> | any => {
		if (fieldName === "Data")	{
            let dataObj = {
				Sender: populateAttribute("String"),
				ID: populateAttribute("String"),
				Content: populateAttribute("String")
			};
			console.log('[ModelOpsCMDConverter::value] dataObj : ', dataObj);
            	return JSON.stringify(dataObj);		
        } else if (fieldName === "Reply")	{
            let replyObj = {
				Sender: populateAttribute("String"),
				Id: populateAttribute("String"),
				Content: populateAttribute("String"),
				Status: populateAttribute("Boolean")
			};
			console.log('[ModelOpsCMDConverter::value] replyObj : ', replyObj);
            	return JSON.stringify(replyObj);		
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