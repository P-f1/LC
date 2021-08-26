"use strict";function buildSchema(e,t,r){return Observable_1.Observable.create(function(n){wi_contrib_1.WiContributionUtils.getConnections(e,"GraphBuilder_Tools").subscribe(function(e){var a;e.forEach(function(e){for(var r,n=0,o=e.settings;n<o.length;n++){var i=o[n];"name"===i.name?r=i.value:"schema"===i.name&&t===r&&(a=i.value)}}),n.next(JSON.stringify(r(a))),n.complete()})})}function populateAttribute(e){switch(e){case"Double":case"Integer":case"Long":return 2;case"Boolean":return!0;case"Date":return 2;case"Object":return{};case"Array":return[];default:return"2"}}var __extends=this&&this.__extends||function(){var e=function(t,r){return(e=Object.setPrototypeOf||{__proto__:[]}instanceof Array&&function(e,t){e.__proto__=t}||function(e,t){for(var r in t)t.hasOwnProperty(r)&&(e[r]=t[r])})(t,r)};return function(t,r){function n(){this.constructor=t}e(t,r),t.prototype=null===r?Object.create(r):(n.prototype=r.prototype,new n)}}(),__decorate=this&&this.__decorate||function(e,t,r,n){var a,o=arguments.length,i=o<3?t:null===n?n=Object.getOwnPropertyDescriptor(t,r):n;if("object"==typeof Reflect&&"function"==typeof Reflect.decorate)i=Reflect.decorate(e,t,r,n);else for(var c=e.length-1;c>=0;c--)(a=e[c])&&(i=(o<3?a(i):o>3?a(t,r,i):a(t,r))||i);return o>3&&i&&Object.defineProperty(t,r,i),i},__metadata=this&&this.__metadata||function(e,t){if("object"==typeof Reflect&&"function"==typeof Reflect.metadata)return Reflect.metadata(e,t)},__param=this&&this.__param||function(e,t){return function(r,n){t(r,n,e)}};Object.defineProperty(exports,"__esModule",{value:!0});var Observable_1=require("rxjs/Observable"),core_1=require("@angular/core"),http_1=require("@angular/http"),wi_contrib_1=require("wi-studio/app/contrib/wi-contrib"),TableMutateContributionHandler=function(e){function t(t,r){var n=e.call(this,t,r)||this;return n.http=r,n.value=function(e,t){if(console.log("[TableContributionHandler::value] fieldName = "+e),"Table"===e){for(var r=t.getField("Table").allowed,a=t.getField("Table").value,o=0,i=r;o<i.length;o++){var c=i[o];c.unique_id===a&&(n.selectedConnector=c.name)}return Observable_1.Observable.create(function(e){wi_contrib_1.WiContributionUtils.getConnections(n.http,"GraphBuilder_Tools").subscribe(function(t){var r=[];t.forEach(function(e){for(var t=0,n=e.settings;t<n.length;t++){var a=n[t];"name"===a.name&&r.push({unique_id:wi_contrib_1.WiContributionUtils.getUniqueId(e),name:a.value})}}),e.next(r),e.complete()})})}return"Mapping"===e?buildSchema(n.http,n.selectedConnector,function(e){var t={};if(e)for(var r=JSON.parse(e),n=0;n<r.length;n++)t[r[n].Name]=populateAttribute(r[n].Type);return t}):"Data"===e?buildSchema(n.http,n.selectedConnector,function(e){var t={};if(e)for(var r=JSON.parse(e),n=0;n<r.length;n++)t[r[n].Name]=populateAttribute(r[n].Type);return{New:t,Old:t}}):null},n.validate=function(e,t){return console.log("[TableContributionHandler::validate] fieldName = "+e),null},n}return __extends(t,e),t=__decorate([wi_contrib_1.WiContrib({}),core_1.Injectable(),__param(0,core_1.Inject(core_1.Injector)),__metadata("design:paramtypes",[Object,http_1.Http])],t)}(wi_contrib_1.WiServiceHandlerContribution);exports.TableMutateContributionHandler=TableMutateContributionHandler;