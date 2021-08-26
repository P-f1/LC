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
export class KubernetesDeployActivityHandler extends WiServiceHandlerContribution {
	selectedConnector: string;
		
    constructor(@Inject(Injector) injector, private http: Http) {
        super(injector, http);
    }

    value = (fieldName: string, context: IActivityContribution): Observable<any> | any => {

		console.log('[ProjAirCMDConverter::value] Build field : ', fieldName);
        if (fieldName === "Components") {
            let componentsObj = [{
	 			"Type": populateAttribute("String"),
				"Name" : populateAttribute("String"),
            		Properties : [{
					Name: populateAttribute("String"),
					Value: populateAttribute("String")
				}]
			}];
		console.log('[ProjAirCMDConverter::value] componentsObj : ', componentsObj);
            	return JSON.stringify(componentsObj);
        } else if (fieldName === "DeployDescriptor") {
            let args = {
            		Command: populateAttribute("String"),
				Deployment: populateAttribute("String"),
				Replicas:   populateAttribute("Integer"),
            		Containers : [{
					Name: populateAttribute("String"),
					Image: populateAttribute("String"),
					Ports: {
						ContainerPort: populateAttribute("Integer")
					},
					Env: [{
						Name: populateAttribute("String"),
						Value: populateAttribute("String")
					}]
				}]
			};
		console.log('[ProjAirCMDConverter::value] K8DeployDescriptor : ', args);
            	return JSON.stringify(args);
        }
		
        return null;
    }
 
    validate = (fieldName: string, context: IActivityContribution): Observable<IValidationResult> | IValidationResult => {
		console.log('[ProjAirCMDConverter::value] Validate field : ', fieldName);

        if (fieldName === "Cluster") {
            let connection: IFieldDefinition = context.getField("Cluster")
        		if (connection.value === null) {
            		return ValidationResult.newValidationResult().setError("KubernetesDeploy-MSG-1000", "Kubernetes deploy must be configured");
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