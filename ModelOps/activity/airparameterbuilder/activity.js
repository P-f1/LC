"use strict";function populateAttribute(t){switch(t){case"Double":case"Integer":case"Long":return 2;case"Boolean":return!0;case"Date":return 2;default:return"2"}}var __extends=this&&this.__extends||function(){var t=function(e,r){return(t=Object.setPrototypeOf||{__proto__:[]}instanceof Array&&function(t,e){t.__proto__=e}||function(t,e){for(var r in e)e.hasOwnProperty(r)&&(t[r]=e[r])})(e,r)};return function(e,r){function i(){this.constructor=e}t(e,r),e.prototype=null===r?Object.create(r):(i.prototype=r.prototype,new i)}}(),__decorate=this&&this.__decorate||function(t,e,r,i){var n,a=arguments.length,o=a<3?e:null===i?i=Object.getOwnPropertyDescriptor(e,r):i;if("object"==typeof Reflect&&"function"==typeof Reflect.decorate)o=Reflect.decorate(t,e,r,i);else for(var u=t.length-1;u>=0;u--)(n=t[u])&&(o=(a<3?n(o):a>3?n(e,r,o):n(e,r))||o);return a>3&&o&&Object.defineProperty(e,r,o),o},__metadata=this&&this.__metadata||function(t,e){if("object"==typeof Reflect&&"function"==typeof Reflect.metadata)return Reflect.metadata(t,e)},__param=this&&this.__param||function(t,e){return function(r,i){e(r,i,t)}};Object.defineProperty(exports,"__esModule",{value:!0});var core_1=require("@angular/core"),http_1=require("@angular/http"),wi_contrib_1=require("wi-studio/app/contrib/wi-contrib"),AirParameterBuilderActivityHandler=function(t){function e(e,r){var i=t.call(this,e,r)||this;return i.http=r,i.value=function(t,e){if(console.log("[AirParameterBuilder::value] Build field : ",t),"Descriptor"===t){var r={FlogoDescriptor:populateAttribute("String"),F1Properties:[{Group:populateAttribute("String"),Value:{Name:populateAttribute("String"),Value:populateAttribute("String"),Type:populateAttribute("String")}}]};return JSON.stringify(r)}if("Variables"===t){var i={},n=e.getField("variablesDef");if(n.value)for(var a=JSON.parse(n.value),o=0;o<a.length;o++)i[a[o].Name]=populateAttribute(a[o].Type);return JSON.stringify(i)}return null},i.validate=function(t,e){return null},i}return __extends(e,t),e=__decorate([wi_contrib_1.WiContrib({}),core_1.Injectable(),__param(0,core_1.Inject(core_1.Injector)),__metadata("design:paramtypes",[Object,http_1.Http])],e)}(wi_contrib_1.WiServiceHandlerContribution);exports.AirParameterBuilderActivityHandler=AirParameterBuilderActivityHandler;