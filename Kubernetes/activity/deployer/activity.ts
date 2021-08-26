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

		console.log('[KubernetesDeploy::value] Build field : ', fieldName);
        if (fieldName === "Context") {
			console.log('\n\n[KubernetesDeploy::value] ================== ', fieldName);
            let allowedConnectors = context.getField("Context").allowed;	
			let selectedConnectorId = context.getField("Context").value;
			for(let allowedConnector of allowedConnectors) {
				if(allowedConnector["unique_id"] === selectedConnectorId) {
					this.selectedConnector = allowedConnector["name"]
				}
			}
            
            return Observable.create(observer => {
            	//Connector Type must match with the category defined in connector.json
                WiContributionUtils.getConnections(this.http, "Kubernetes").subscribe((data: IConnectorContribution[]) => {
                		let connectionRefs = [];
                    data.forEach(connection => {
                        for (let setting of connection.settings) {
							if(setting.name === "name") {
								connectionRefs.push({
									"unique_id": WiContributionUtils.getUniqueId(connection),
									"name": setting.value
								});
							}
                        }
                    });
                    observer.next(connectionRefs);
                		observer.complete();
                });
            });
        } else if (fieldName === "Deployplan") {
            let deployplanObj = {
            		Strategy: populateAttribute("String"),
            		Pipeline : [{
					GroupID : 	populateAttribute("String"),
					Deployments :[{
						Type : populateAttribute("String"),
						Name : populateAttribute("String")
					}]
				}]
			};
			console.log('[KubernetesDeployActivityHandler::value] deployplanObj : ', deployplanObj);
            return JSON.stringify(deployplanObj);
        } else if (fieldName === "Deployments") {
            let deploymentsArr = [{
            	Command: populateAttribute("String"),
				Deployment: populateAttribute("String"),
				IsEndpoint: populateAttribute("Boolean"),
				Replicas:   populateAttribute("Integer"),
            	Containers : [{
					Name: populateAttribute("String"),
					Image: populateAttribute("String"),
					Ports: {
						ContainerPort: populateAttribute("Integer")
					},
					VolumeMounts:[{
						MountPath: populateAttribute("String"),
						Name: populateAttribute("String")
					}],
					Env: [{
						Name: populateAttribute("String"),
						Value: populateAttribute("String")
					}]
				}],
				Volumes: [{
					Name: populateAttribute("String"),
					HostPath: {
						Path: populateAttribute("String")
					}
				}]
			}];
		console.log('[KubernetesDeployActivityHandler::value] deploymentsArr : ', deploymentsArr);
            	return JSON.stringify(deploymentsArr);
        }
		
        return null;
    }
 
    validate = (fieldName: string, context: IActivityContribution): Observable<IValidationResult> | IValidationResult => {
		console.log('[GraphBuilder::value] Validate field : ', fieldName);

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