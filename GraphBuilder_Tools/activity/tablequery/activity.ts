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
export class TableQueryContributionHandler extends WiServiceHandlerContribution {
	selectedConnector: string;
	selectedIndices: string;

    constructor(@Inject(Injector) injector, private http: Http) {
        super(injector, http);
    }

    value = (fieldName: string, context: IActivityContribution): Observable<any> | any => {
    	
		console.log('[TableQueryContributionHandler::value] fieldName = ' + fieldName);
		let indices = [];
		let indicesDataType = {};

        if (fieldName === "Table") {
			console.log('[TableQueryContributionHandler::value] in Table !!');
			let allowedConnectors = context.getField("Table").allowed;	
			let selectedConnectorId = context.getField("Table").value;
			for(let allowedConnector of allowedConnectors) {
				if(allowedConnector["unique_id"] === selectedConnectorId) {
					this.selectedConnector = allowedConnector["name"]
				}
			}

            return Observable.create(observer => {
            	//Connector Type must match with the category defined in connector.json
                WiContributionUtils.getConnections(this.http, "GraphBuilder_Tools").subscribe((data: IConnectorContribution[]) => {
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
        } else if (fieldName === "Indices") {
			this.selectedIndices = context.getField("Indices").value;
			console.log('[TableQueryContributionHandler::value] in Indices !!');
        	return buildSchema(this.http, this.selectedConnector, (schema : string) => {
            	var indices = [];
            	if (schema) {
					/* Build a list of elements which can be indexed */
					var indexElements = [];
                	let data = JSON.parse(schema);
                	for (var i = 0; i < data.length; i++) {
						if(data[i].IsKey==="yes"||data[i].Indexed==="yes") {
							indexElements.push(data[i].Name);
						}
						indicesDataType[data[i].Name] = populateAttribute(data[i].Type);
                	}
					
					/* Build all possible querykey */
					var combs = [];
					for (var j=1; j<indexElements.length+1; j++) {
						data = [];
						data.length = j;
						combination(indexElements, data, 0, indexElements.length-1, 0, j, combs);
					}
					for (var k=0; k<combs.length; k++) {				
						indices.push({
							"unique_id": combs[k],
							"name": combs[k]
						})
					}
            	}
            	return indices;
			});
        } else if (fieldName === "QueryKey") {
			console.log('[TableQueryContributionHandler::value] in QueryKey !!');
			var attrJsonKey = {};
			let data = this.selectedIndices.split(" ");
			for (var i = 0; i < data.length; i++) {
				attrJsonKey[data[i]] = populateAttribute(indicesDataType[data[i]]);
			}
			return JSON.stringify(attrJsonKey);
        } else if (fieldName === "Data") {	
			console.log('[TableQueryContributionHandler::value] in Data !!');
        	return buildSchema(this.http, this.selectedConnector, (schema : string) => {
            	if (schema) {
					var primaryKeyElements = [];
					var indexElements = [];
                	let data = JSON.parse(schema);
                	for (var i = 0; i < data.length; i++) {
						if(data[i].IsKey==="yes") {
							primaryKeyElements.push(data[i].Name);
						}
                	}
					
					/* Build promarykey */
					var primaryKey = ""
        			for (var i=0; i<primaryKeyElements.length; i++) {
            			if(i!==0)
            				primaryKey += " ";
            			primaryKey += primaryKeyElements[i];
        			} 
					
					if (primaryKey==this.selectedIndices) {
						let attrJsonSchema = {};
                		for (var i = 0; i < data.length; i++) {
                			attrJsonSchema[data[i].Name] = populateAttribute(data[i].Type);
                		}
            			return JSON.stringify(attrJsonSchema);					
					} else {
						let attrJsonSchema = [{}];
                		for (var i = 0; i < data.length; i++) {
                			attrJsonSchema[0][data[i].Name] = populateAttribute(data[i].Type);
                		}
            			return JSON.stringify(attrJsonSchema);
					}
            	}
            	return JSON.stringify({});
			});
        }
        
        return null;
    }
 
    validate = (fieldName: string, context: IActivityContribution): Observable<IValidationResult> | IValidationResult => {
		
		console.log('[TableContributionHandler::validate] fieldName = ' + fieldName);

		return null; 
    }
}

function buildSchema(http, selectedConnector, builder) : Observable<any> {
	return Observable.create(observer => {
		WiContributionUtils.getConnections(http, "GraphBuilder_Tools").subscribe((data: IConnectorContribution[]) => {
		var schema;
        data.forEach(connection => {
			var currentConnector;
           for (let setting of connection.settings) {
				if(setting.name === "name") {
					currentConnector = setting.value
				}else if (setting.name === "schema"&&
					selectedConnector === currentConnector) {
						schema = setting.value;
                	}
            	}
        	});
			console.log("*****************************************************")
			console.log(builder(schema))
			console.log("*****************************************************")
			observer.next(builder(schema));
			observer.complete();
		});
	});			
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
		case "Object":
			return {};
		case "Array":
			return [];
		default:
    		return "2";
	}
}

function combination(arr, data, start, end, index, r, combs) {     
    if (index === r) {
        var comb = "";
        for (var j=0; j<r; j++) {
            if(j!==0)
            	comb += " ";
            comb += data[j];
        } 
        combs.push(comb);
        return; 
    }
  
    var i = start;  
    while(i <= end && end - i + 1 >= r - index){ 
        data[index] = arr[i]; 
        combination(arr, data, i + 1, end, index + 1, r, combs);
        i += 1; 
    }
}