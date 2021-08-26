"use strict";function populateAttribute(t){switch(t){case"Double":case"Integer":case"Long":return 2;case"Boolean":return!0;case"Date":return 2;default:return"2"}}var __extends=this&&this.__extends||function(){var t=function(e,r){return(t=Object.setPrototypeOf||{__proto__:[]}instanceof Array&&function(t,e){t.__proto__=e}||function(t,e){for(var r in e)e.hasOwnProperty(r)&&(t[r]=e[r])})(e,r)};return function(e,r){function n(){this.constructor=e}t(e,r),e.prototype=null===r?Object.create(r):(n.prototype=r.prototype,new n)}}(),__decorate=this&&this.__decorate||function(t,e,r,n){var o,a=arguments.length,i=a<3?e:null===n?n=Object.getOwnPropertyDescriptor(e,r):n;if("object"==typeof Reflect&&"function"==typeof Reflect.decorate)i=Reflect.decorate(t,e,r,n);else for(var c=t.length-1;c>=0;c--)(o=t[c])&&(i=(a<3?o(i):a>3?o(e,r,i):o(e,r))||i);return a>3&&i&&Object.defineProperty(e,r,i),i},__metadata=this&&this.__metadata||function(t,e){if("object"==typeof Reflect&&"function"==typeof Reflect.metadata)return Reflect.metadata(t,e)},__param=this&&this.__param||function(t,e){return function(r,n){e(r,n,t)}};Object.defineProperty(exports,"__esModule",{value:!0});var core_1=require("@angular/core"),http_1=require("@angular/http"),wi_contrib_1=require("wi-studio/app/contrib/wi-contrib"),JSONDataDecouplerContributionHandler=function(t){function e(e,r){var n=t.call(this,e,r)||this;return n.http=r,n.value=function(t,e){console.log("fieldName = "+t);var r;if("JSONObject"===t){for(var n=0,o=e.settings;n<o.length;n++){if("sample"===(c=o[n]).name){if(c,c.value)return c.value.filename,(r=c.value.content)&&(r=r.substr(r.indexOf(",")+1),r=atob(r)),JSON.stringify(JSON.parse(r))}}return JSON.stringify({})}if("Data"===t){for(var a=0,i=e.settings;a<i.length;a++){var c;if("sample"===(c=i[a]).name){if(c,c.value){console.log(e.settings),c.value.filename,(r=c.value.content)&&(r=r.substr(r.indexOf(",")+1),r=atob(r));var u=[{}],l=JSON.parse(r),s=e.getField("decoupleTarget").value;console.log(s),u[0].originJSONObject=l;var f=l,p=s.split(".");for(var _ in p)f=f[p[_]];return u[0][s.concat(".Index")]=populateAttribute("Integer"),u[0][s.concat(".Element")]=f[0],u[0].LastElement=populateAttribute("Boolean"),JSON.stringify(u)}}}return JSON.stringify({})}return null},n.validate=function(t,e){return null},n}return __extends(e,t),e=__decorate([wi_contrib_1.WiContrib({}),core_1.Injectable(),__param(0,core_1.Inject(core_1.Injector)),__metadata("design:paramtypes",[Object,http_1.Http])],e)}(wi_contrib_1.WiServiceHandlerContribution);exports.JSONDataDecouplerContributionHandler=JSONDataDecouplerContributionHandler;