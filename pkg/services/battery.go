/*
 * Copyright 2022 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package services

import (
	"github.com/SENERGY-Platform/mgw-dc-lib-go/pkg/mgw"
	"github.com/SENERGY-Platform/mgw-ssh-dc/pkg/model"
	"strconv"
	"strings"
	"time"
)

func init() {
	d := 30 * time.Second
	Services = append(Services, mgw.Service[*model.SshDcDevice]{
		F:       GetBattery,
		D:       &d,
		LocalId: "battery",
	})
}

func GetBattery(device *model.SshDcDevice, _ interface{}) (result interface{}, err error) {
	bytes, err := device.SshClient.Run("cat /sys/class/power_supply/BAT0/capacity")
	if err != nil {
		// probably file not found
		return nil, nil
	}
	s := strings.TrimSuffix(string(bytes), "\n")
	result, err = strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	return
}
