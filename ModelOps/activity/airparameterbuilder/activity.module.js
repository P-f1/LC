"use strict";var __decorate=this&&this.__decorate||function(e,r,t,i){var o,c=arguments.length,n=c<3?r:null===i?i=Object.getOwnPropertyDescriptor(r,t):i;if("object"==typeof Reflect&&"function"==typeof Reflect.decorate)n=Reflect.decorate(e,r,t,i);else for(var a=e.length-1;a>=0;a--)(o=e[a])&&(n=(c<3?o(n):c>3?o(r,t,n):o(r,t))||n);return c>3&&n&&Object.defineProperty(r,t,n),n};Object.defineProperty(exports,"__esModule",{value:!0});var http_1=require("@angular/http"),core_1=require("@angular/core"),common_1=require("@angular/common"),activity_1=require("./activity"),wi_contrib_1=require("wi-studio/app/contrib/wi-contrib"),AirParameterBuilderActivityHandlerImpl=function(){function e(){}return e=__decorate([core_1.NgModule({imports:[common_1.CommonModule,http_1.HttpModule],providers:[{provide:wi_contrib_1.WiServiceContribution,useClass:activity_1.AirParameterBuilderActivityHandler}]})],e)}();exports.default=AirParameterBuilderActivityHandlerImpl;