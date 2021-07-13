// Copyright 2019 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"reflect"
	"time"

	"github.com/golang/glog"
	"github.com/kubevirt/bridge-marker/pkg/cache"
	"github.com/kubevirt/bridge-marker/pkg/marker"
	"k8s.io/apimachinery/pkg/util/wait"
)

func main() {
	nodeName := flag.String("node-name", "", "name of kubernetes node")
	const defaultUpdateInterval = 60 * time.Second
	updateInterval := flag.Int("update-interval", int(defaultUpdateInterval.Seconds()), fmt.Sprintf("interval between updates in seconds, %d by default", defaultUpdateInterval))

	const defaultReconcileInterval = 10 * time.Minute
	reconcileInterval := flag.Int("reconcile-interval", int(defaultReconcileInterval.Minutes()), fmt.Sprintf("interval between node bridges reconcile in minutes, %d by default", defaultReconcileInterval))

	flag.Parse()

	if *nodeName == "" {
		glog.Fatal("node-name must be set")
	}

	cache := cache.Cache{}
	wait.PollImmediateInfinite(time.Duration(*updateInterval)*time.Second, func() (bool, error) {
		if time.Now().Sub(cache.LastRefreshTime()) >= time.Duration(*reconcileInterval)*time.Minute {
			reportedBridges, err := marker.GetReportedResources(*nodeName)
			if err != nil {
				glog.Errorf("GetReportedResources failed: %v", err)
			}

			if !reflect.DeepEqual(cache.Bridges(), reportedBridges) {
				glog.Warningf("cached bridges are different than the reported bridges on node %s", *nodeName)
			}
			
			cache.Refresh(reportedBridges)
		}

		err := marker.Update(*nodeName, cache)
		if err != nil {
			glog.Errorf("Update failed: %v", err)
		}

		return false, nil
	})
}
