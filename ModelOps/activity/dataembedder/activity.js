"use strict";function populateAttribute(e){switch(e){case"Double":case"Integer":case"Long":return 2;case"Boolean":return!0;case"Date":return 2;default:return"2"}}var __extends=this&&this.__extends||function(){var e=function(t,r){return(e=Object.setPrototypeOf||{__proto__:[]}instanceof Array&&function(e,t){e.__proto__=t}||function(e,t){for(var r in t)t.hasOwnProperty(r)&&(e[r]=t[r])})(t,r)};return function(t,r){function n(){this.constructor=t}e(t,r),t.prototype=null===r?Object.create(r):(n.prototype=r.prototype,new n)}}(),__decorate=this&&this.__decorate||function(e,t,r,n){var i,o=arguments.length,a=o<3?t:null===n?n=Object.getOwnPropertyDescriptor(t,r):n;if("object"==typeof Reflect&&"function"==typeof Reflect.decorate)a=Reflect.decorate(e,t,r,n);else for(var c=e.length-1;c>=0;c--)(i=e[c])&&(a=(o<3?i(a):o>3?i(t,r,a):i(t,r))||a);return o>3&&a&&Object.defineProperty(t,r,a),a},__metadata=this&&this.__metadata||function(e,t){if("object"==typeof Reflect&&"function"==typeof Reflect.metadata)return Reflect.metadata(e,t)},__param=this&&this.__param||function(e,t){return function(r,n){t(r,n,e)}};Object.defineProperty(exports,"__esModule",{value:!0});var core_1=require("@angular/core"),http_1=require("@angular/http"),wi_contrib_1=require("wi-studio/app/contrib/wi-contrib"),AirPipelineBuilderActivityHandler=function(e){function t(t,r){var n=e.call(this,t,r)||this;return n.http=r,n.value=function(e,t){if(console.log("[AirPipelineBuilder::value] Build field : ",e),"TargetData"===e){var r={},n=t.getField("Targets");if(n.value)for(var i=JSON.parse(n.value),o=0;o<i.length;o++)r[i[o].Name]=populateAttribute(i[o].Type);return JSON.stringify(r)}return null},n.validate=function(e,t){return null},n}return __extends(t,e),t=__decorate([wi_contrib_1.WiContrib({}),core_1.Injectable(),__param(0,core_1.Inject(core_1.Injector)),__metadata("design:paramtypes",[Object,http_1.Http])],t)}(wi_contrib_1.WiServiceHandlerContribution);exports.AirPipelineBuilderActivityHandler=AirPipelineBuilderActivityHandler;