/*
 * Copyright © 2020. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
import { HttpModule } from "@angular/http";
import { NgModule } from "@angular/core";
import { CommonModule } from "@angular/common";
import { Descriptor2K8sYamlActivityyHandler} from "./activity";
import { WiServiceContribution} from "wi-studio/app/contrib/wi-contrib";

@NgModule({
  imports: [
    CommonModule,
    HttpModule,
 ],
  providers: [
    {
       provide: WiServiceContribution,
       useClass: Descriptor2K8sYamlActivityyHandler
     }
  ]
})

export default class Descriptor2K8sYamlActivityyHandlerImpl {

}