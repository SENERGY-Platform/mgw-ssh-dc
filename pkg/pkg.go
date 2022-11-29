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

package pkg

import (
	"context"
	"github.com/SENERGY-Platform/mgw-dc-lib-go/pkg/configuration"
	"github.com/SENERGY-Platform/mgw-dc-lib-go/pkg/mgw"
	"github.com/SENERGY-Platform/mgw-ssh-dc/pkg/model"
	"github.com/SENERGY-Platform/mgw-ssh-dc/pkg/services"
	"sync"
	"time"
)

type SshDc struct {
	config    Config
	libConfig configuration.Config
	client    *mgw.Client[*model.SshDcDevice]
}

const dcPrefix = "mgw-ssh-dc:"

func Start(config Config, libConfig configuration.Config, ctx context.Context, wg *sync.WaitGroup) (err error) {
	dc := &SshDc{
		config:    config,
		libConfig: libConfig,
	}
	client, err := mgw.New[*model.SshDcDevice](libConfig, ctx, wg, dc.Discover)
	if err != nil {
		return err
	}
	dc.client = client
	for _, s := range services.Services {
		client.RegisterServiceStruct(s)
	}
	go func() {
		t := time.NewTicker(time.Minute)
		for {
			select {
			case <-t.C:
				dc.Discover()
				break
			case <-ctx.Done():
				return
			}
		}
	}()
	return
}
