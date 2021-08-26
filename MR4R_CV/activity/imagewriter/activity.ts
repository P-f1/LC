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
export class ImageProcessorContributionHandler extends WiServiceHandlerContribution {
	filename: string;
	content: string; 

    constructor(@Inject(Injector) injector, private http: Http) {
        super(injector, http);
    }

    value = (fieldName: string, context: IActivityContribution): Observable<any> | any => {
        if (fieldName === "Connection") {
            //Connector Type must match with the category defined in connector.json
            return Observable.create(observer => {
                let connectionRefs = [];                
                WiContributionUtils.getConnections(this.http, "MR4R_CV").subscribe((data: IConnectorContribution[]) => {
                    data.forEach(connection => {
						let connector: string;
                        for (let i = 0; i < connection.settings.length; i++) {
                        		if(connection.settings[i].name === "name") {
								connector = connection.settings[i].value
							}
                        }
                        	connectionRefs.push({
                        		"unique_id": WiContributionUtils.getUniqueId(connection),
                        		"name": connector
                        	});
                    });
                    observer.next(connectionRefs);
                });
            });
        } else if (fieldName === "Commands") {
            let args = [{
            		Command:"Up",
				Argument:{speed:25}
            }];
            	return JSON.stringify(args);
        }
        
        return null;
    }
     
    validate = (fieldName: string, context: IActivityContribution): Observable<IValidationResult> | IValidationResult => {
		
		console.log('validate >>>>>>>> fieldName = ' + fieldName);

        if (fieldName === "Connection") {
            let connection: IFieldDefinition = context.getField("Connection")
        		if (connection.value === null) {
        			return ValidationResult.newValidationResult().setError("TELLO-CMD-SENDER-MSG-1000", "Connector must be configured");
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