/*
 * Copyright © 2020. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package modelrunner

import (
	"sync"

	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/logger"
)

var log = logger.GetLogger("activity-modelrunner")

type Mapping struct {
	metadata *activity.Metadata
	mux      sync.Mutex
}

func NewActivity(metadata *activity.Metadata) activity.Activity {
	activity := &Mapping{metadata: metadata}
	return activity
}

func (a *Mapping) Metadata() *activity.Metadata {
	return a.metadata
}

func (a *Mapping) Eval(ctx activity.Context) (done bool, err error) {

	err = a.init(ctx)

	if nil != err {
		return false, err
	}

	return true, nil
}

func (a *Mapping) init(context activity.Context) error {

	return nil
}
