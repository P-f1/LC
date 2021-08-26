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
export class DockerImageBuilderActivityHandler extends WiServiceHandlerContribution {
	selectedConnector: string;
		
    constructor(@Inject(Injector) injector, private http: Http) {
        super(injector, http);
    }

    value = (fieldName: string, context: IActivityContribution): Observable<any> | any => {		
        return null;
    }
 
    validate = (fieldName: string, context: IActivityContribution): Observable<IValidationResult> | IValidationResult => {
		console.log('[DockerImageBuilderActivityHandler::validate] Validate field : ', fieldName);
		let useRESTful: IFieldDefinition = context.getField("UseRESTful")
		let command: IFieldDefinition = context.getField("Command")
		console.log('[GraphBuilder::value] useRESTful : ', useRESTful);
			
		if (fieldName === "DockerHost") {
			if (useRESTful.value === true) {
				return ValidationResult.newValidationResult().setVisible(true);
			} else {
				return ValidationResult.newValidationResult().setVisible(false);
			}
		} else if (fieldName === "DockerSocket") {
			if (useRESTful.value === false) {
				return ValidationResult.newValidationResult().setVisible(true);
			} else {
				return ValidationResult.newValidationResult().setVisible(false);
			}
		} else if (fieldName === "DockerFile") {
			if (command.value === "push") {
				return ValidationResult.newValidationResult().setVisible(false);
			} else {
				return ValidationResult.newValidationResult().setVisible(true);
			}
		} else if (fieldName === "WorkingFolder") {
			if (command.value === "push") {
				return ValidationResult.newValidationResult().setVisible(false);
			} else {
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