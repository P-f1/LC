/*
 * Copyright © 2017. TIBCO Software Inc.
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
export class VideoReceiverContributionHandler extends WiServiceHandlerContribution {
	url: string;
	user: string; 
	passsword: string; 
		
    constructor(@Inject(Injector) injector, private http: Http) {
        super(injector, http);
    }

    value = (fieldName: string, context: IActivityContribution): Observable<any> | any => {
        if (fieldName === "Tello") {
            //Connector Type must match with the category defined in connector.json
            return Observable.create(observer => {
                let connectionRefs = [];                
                WiContributionUtils.getConnections(this.http, "MR4R_TELLO").subscribe((data: IConnectorContribution[]) => {
                    data.forEach(connection => {
						let connector: string;
						let outbound: boolean; 
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
        }
        
        return null;
    }
 
    validate = (fieldName: string, context: IActivityContribution): Observable<IValidationResult> | IValidationResult => {
		console.log('validate >>>>>>>> fieldName = ' + fieldName);
		if (fieldName === "FramePerSec") {			
            let rawImage: IFieldDefinition = context.getField("RawImage")
			console.log('[ImageReader::value] rawImage : ', rawImage);
        		if (rawImage.value === true) {
			console.log('[ImageReader::value] rawImage.value : ', rawImage.value);
            		return ValidationResult.newValidationResult().setVisible(false);
        		} else {
			console.log('[ImageReader::value] rawImage.value : ', rawImage.value);
				return ValidationResult.newValidationResult().setVisible(true);
			}
        }
        
        return null; 
    }
}