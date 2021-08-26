"use strict";function populateAttribute(t){switch(t){case"Double":case"Integer":case"Long":return 2;case"Boolean":return!0;case"Date":return 2;default:return"2"}}var __extends=this&&this.__extends||function(){var t=function(e,n){return(t=Object.setPrototypeOf||{__proto__:[]}instanceof Array&&function(t,e){t.__proto__=e}||function(t,e){for(var n in e)e.hasOwnProperty(n)&&(t[n]=e[n])})(e,n)};return function(e,n){function r(){this.constructor=e}t(e,n),e.prototype=null===n?Object.create(n):(r.prototype=n.prototype,new r)}}(),__decorate=this&&this.__decorate||function(t,e,n,r){var o,i=arguments.length,a=i<3?e:null===r?r=Object.getOwnPropertyDescriptor(e,n):r;if("object"==typeof Reflect&&"function"==typeof Reflect.decorate)a=Reflect.decorate(t,e,n,r);else for(var c=t.length-1;c>=0;c--)(o=t[c])&&(a=(i<3?o(a):i>3?o(e,n,a):o(e,n))||a);return i>3&&a&&Object.defineProperty(e,n,a),a},__metadata=this&&this.__metadata||function(t,e){if("object"==typeof Reflect&&"function"==typeof Reflect.metadata)return Reflect.metadata(t,e)},__param=this&&this.__param||function(t,e){return function(n,r){e(n,r,t)}};Object.defineProperty(exports,"__esModule",{value:!0});var Observable_1=require("rxjs/Observable"),core_1=require("@angular/core"),http_1=require("@angular/http"),wi_contrib_1=require("wi-studio/app/contrib/wi-contrib"),VideoPlayerContributionHandler=function(t){function e(e,n){var r=t.call(this,e,n)||this;return r.http=n,r.value=function(t,e){if("Connection"===t)return Observable_1.Observable.create(function(t){var e=[];wi_contrib_1.WiContributionUtils.getConnections(r.http,"MR4R_CV").subscribe(function(n){n.forEach(function(t){for(var n,r=0;r<t.settings.length;r++)"name"===t.settings[r].name&&(n=t.settings[r].value);e.push({unique_id:wi_contrib_1.WiContributionUtils.getUniqueId(t),name:n})}),t.next(e)})});if("Commands"===t){var n=[{Command:"Up",Argument:{speed:25}}];return JSON.stringify(n)}return null},r.validate=function(t,e){if(console.log("validate >>>>>>>> fieldName = "+t),"Connection"===t){if(null===e.getField("Connection").value)return wi_contrib_1.ValidationResult.newValidationResult().setError("TELLO-CMD-SENDER-MSG-1000","Connector must be configured")}return null},r}return __extends(e,t),e=__decorate([wi_contrib_1.WiContrib({}),core_1.Injectable(),__param(0,core_1.Inject(core_1.Injector)),__metadata("design:paramtypes",[Object,http_1.Http])],e)}(wi_contrib_1.WiServiceHandlerContribution);exports.VideoPlayerContributionHandler=VideoPlayerContributionHandler;