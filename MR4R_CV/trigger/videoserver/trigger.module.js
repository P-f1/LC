"use strict";var __decorate=this&&this.__decorate||function(e,r,t,o){var i,n=arguments.length,c=n<3?r:null===o?o=Object.getOwnPropertyDescriptor(r,t):o;if("object"==typeof Reflect&&"function"==typeof Reflect.decorate)c=Reflect.decorate(e,r,t,o);else for(var u=e.length-1;u>=0;u--)(i=e[u])&&(c=(n<3?i(c):n>3?i(r,t,c):i(r,t))||c);return n>3&&c&&Object.defineProperty(r,t,c),c};Object.defineProperty(exports,"__esModule",{value:!0});var http_1=require("@angular/http"),core_1=require("@angular/core"),common_1=require("@angular/common"),trigger_1=require("./trigger"),wi_contrib_1=require("wi-studio/app/contrib/wi-contrib"),VideoServerModule=function(){function e(){}return e=__decorate([core_1.NgModule({imports:[common_1.CommonModule,http_1.HttpModule],providers:[{provide:wi_contrib_1.WiServiceContribution,useClass:trigger_1.VideoServerContributionHandler}]})],e)}();exports.default=VideoServerModule;