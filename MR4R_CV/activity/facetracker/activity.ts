/*
 * Copyright © 2019. TIBCO Software Inc.
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
export class FaceTrackerContributionHandler extends WiServiceHandlerContribution {
	filename: string;
	content: string; 

    constructor(@Inject(Injector) injector, private http: Http) {
        super(injector, http);
    }

    value = (fieldName: string, context: IActivityContribution): Observable<any> | any => {
        if (fieldName === "Detection") {
            let args = {
				frame: {
					image: populateAttribute("String"),
					height: populateAttribute("Double"),
					width:  populateAttribute("Double")
				},
				location: {
					left:     populateAttribute("Double"),
					top:      populateAttribute("Double"),
					right:    populateAttribute("Double"),
					bottom:   populateAttribute("Double"),
					distance: populateAttribute("Double")
				},
				delta: {
					x: populateAttribute("Double"),
					y: populateAttribute("Double"),
					z: populateAttribute("Double")
				}
			};
            	return JSON.stringify(args);
        }
        
        return null;
    }
     
    validate = (fieldName: string, context: IActivityContribution): Observable<IValidationResult> | IValidationResult => {
		
		console.log('validate >>>>>>>> fieldName = ' + fieldName);
		console.log(context);
		
		let outputImage: IFieldDefinition = context.getField("OutputImage")

			console.log('outputImage.value01 >>>>>>>> ' + outputImage.value);
		if( fieldName === 'Extension') {
			console.log('outputImage.value02 >>>>>>>> ' + outputImage.value);
        		if (outputImage.value !== true) {
			console.log('outputImage.not visible >>>>>>>> ' + outputImage.value);
				return ValidationResult.newValidationResult().setVisible(false);
			} else {
			console.log('outputImage.visible >>>>>>>> ' + outputImage.value);
				return ValidationResult.newValidationResult().setVisible(true);
			}
		}
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