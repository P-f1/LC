"use strict";var __decorate=this&&this.__decorate||function(e,r,o,t){var c,i=arguments.length,n=i<3?r:null===t?t=Object.getOwnPropertyDescriptor(r,o):t;if("object"==typeof Reflect&&"function"==typeof Reflect.decorate)n=Reflect.decorate(e,r,o,t);else for(var u=e.length-1;u>=0;u--)(c=e[u])&&(n=(i<3?c(n):i>3?c(r,o,n):c(r,o))||n);return i>3&&n&&Object.defineProperty(r,o,n),n};Object.defineProperty(exports,"__esModule",{value:!0});var http_1=require("@angular/http"),core_1=require("@angular/core"),common_1=require("@angular/common"),connector_1=require("./connector"),wi_contrib_1=require("wi-studio/app/contrib/wi-contrib"),TibcoGraphModule=function(){function e(){}return e=__decorate([core_1.NgModule({imports:[common_1.CommonModule,http_1.HttpModule],providers:[{provide:wi_contrib_1.WiServiceContribution,useClass:connector_1.TibcoGraphContribution}]})],e)}();exports.default=TibcoGraphModule;