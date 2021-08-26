"use strict";var __extends=this&&this.__extends||function(){var e=function(t,r){return(e=Object.setPrototypeOf||{__proto__:[]}instanceof Array&&function(e,t){e.__proto__=t}||function(e,t){for(var r in t)t.hasOwnProperty(r)&&(e[r]=t[r])})(t,r)};return function(t,r){function n(){this.constructor=t}e(t,r),t.prototype=null===r?Object.create(r):(n.prototype=r.prototype,new n)}}(),__decorate=this&&this.__decorate||function(e,t,r,n){var i,o=arguments.length,a=o<3?t:null===n?n=Object.getOwnPropertyDescriptor(t,r):n;if("object"==typeof Reflect&&"function"==typeof Reflect.decorate)a=Reflect.decorate(e,t,r,n);else for(var l=e.length-1;l>=0;l--)(i=e[l])&&(a=(o<3?i(a):o>3?i(t,r,a):i(t,r))||a);return o>3&&a&&Object.defineProperty(t,r,a),a},__metadata=this&&this.__metadata||function(e,t){if("object"==typeof Reflect&&"function"==typeof Reflect.metadata)return Reflect.metadata(e,t)},__param=this&&this.__param||function(e,t){return function(r,n){t(r,n,e)}};Object.defineProperty(exports,"__esModule",{value:!0});var Observable_1=require("rxjs/Observable"),core_1=require("@angular/core"),http_1=require("@angular/http"),wi_contrib_1=require("wi-studio/app/contrib/wi-contrib"),YMALDeployActivityHandler=function(e){function t(t,r){var n=e.call(this,t,r)||this;return n.http=r,n.value=function(e,t){if(console.log("[YMALDeploy::value] Build field : ",e),"Cluster"===e){for(var r=t.getField("Cluster").allowed,i=t.getField("Cluster").value,o=0,a=r;o<a.length;o++){var l=a[o];l.unique_id===i&&(n.selectedConnector=l.name)}return Observable_1.Observable.create(function(e){wi_contrib_1.WiContributionUtils.getConnections(n.http,"Kubernetes").subscribe(function(t){var r=[];t.forEach(function(e){for(var t=0,n=e.settings;t<n.length;t++){var i=n[t];"name"===i.name&&r.push({unique_id:wi_contrib_1.WiContributionUtils.getUniqueId(e),name:i.value})}}),e.next(r),e.complete()})})}return null},n.validate=function(e,t){if(console.log("[GraphBuilder::value] Validate field : ",e),"Cluster"===e){if(null===t.getField("Cluster").value)return wi_contrib_1.ValidationResult.newValidationResult().setError("YMALDeploy-MSG-1000","Kubernetes deploy must be configured")}return null},n}return __extends(t,e),t=__decorate([wi_contrib_1.WiContrib({}),core_1.Injectable(),__param(0,core_1.Inject(core_1.Injector)),__metadata("design:paramtypes",[Object,http_1.Http])],t)}(wi_contrib_1.WiServiceHandlerContribution);exports.YMALDeployActivityHandler=YMALDeployActivityHandler;