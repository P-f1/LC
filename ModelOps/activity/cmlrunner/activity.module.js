"use strict";var __decorate=this&&this.__decorate||function(e,t,r,o){var i,n=arguments.length,c=n<3?t:null===o?o=Object.getOwnPropertyDescriptor(t,r):o;if("object"==typeof Reflect&&"function"==typeof Reflect.decorate)c=Reflect.decorate(e,t,r,o);else for(var u=e.length-1;u>=0;u--)(i=e[u])&&(c=(n<3?i(c):n>3?i(t,r,c):i(t,r))||c);return n>3&&c&&Object.defineProperty(t,r,c),c};Object.defineProperty(exports,"__esModule",{value:!0});var http_1=require("@angular/http"),core_1=require("@angular/core"),common_1=require("@angular/common"),activity_1=require("./activity"),wi_contrib_1=require("wi-studio/app/contrib/wi-contrib"),CMLRunnerActivityModule=function(){function e(){}return e=__decorate([core_1.NgModule({imports:[common_1.CommonModule,http_1.HttpModule],providers:[{provide:wi_contrib_1.WiServiceContribution,useClass:activity_1.CMLRunnerActivityContributionHandler}]})],e)}();exports.default=CMLRunnerActivityModule;