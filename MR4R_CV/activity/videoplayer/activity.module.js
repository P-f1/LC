"use strict";var __decorate=this&&this.__decorate||function(e,t,r,o){var i,c=arguments.length,n=c<3?t:null===o?o=Object.getOwnPropertyDescriptor(t,r):o;if("object"==typeof Reflect&&"function"==typeof Reflect.decorate)n=Reflect.decorate(e,t,r,o);else for(var u=e.length-1;u>=0;u--)(i=e[u])&&(n=(c<3?i(n):c>3?i(t,r,n):i(t,r))||n);return c>3&&n&&Object.defineProperty(t,r,n),n};Object.defineProperty(exports,"__esModule",{value:!0});var http_1=require("@angular/http"),core_1=require("@angular/core"),common_1=require("@angular/common"),activity_1=require("./activity"),wi_contrib_1=require("wi-studio/app/contrib/wi-contrib"),VideoPlayerActivityModule=function(){function e(){}return e=__decorate([core_1.NgModule({imports:[common_1.CommonModule,http_1.HttpModule],providers:[{provide:wi_contrib_1.WiServiceContribution,useClass:activity_1.VideoPlayerContributionHandler}]})],e)}();exports.default=VideoPlayerActivityModule;