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
	"github.com/SENERGY-Platform/mgw-dc-lib-go/pkg/mgw"
	"github.com/SENERGY-Platform/mgw-ssh-dc/pkg/model"
	"github.com/melbahja/goph"
	"golang.org/x/exp/slices"
	"log"
	"os"
	"strings"
	"sync"
)

func (dc *SshDc) Discover() {
	if dc.client == nil {
		log.Println("WARN: client not initialized")
		return
	}
	devices := dc.client.GetDevices()
	home := os.Getenv("HOME")
	auth, err := goph.Key(home+"/.ssh/id_rsa", dc.config.SshKeyPw)
	if err != nil {
		errMsg := "cant get ssh key: " + err.Error()
		log.Println(errMsg)
		dc.client.SendClientError(errMsg)
		return
	}
	if len(dc.config.Users) != len(dc.config.Hosts) {
		panic("expected users and ") // TODO
	}
	wg := sync.WaitGroup{}
	updatedIds := []string{}
	updateIdsMux := sync.Mutex{}
	wg.Add(len(dc.config.Hosts))
	for i := range dc.config.Hosts {
		i := i
		go func() {
			defer wg.Done()
			sshClient, err := goph.New(dc.config.Users[i], dc.config.Hosts[i], auth)
			if err != nil {
				log.Println("INFO: Could not connect to host " + dc.config.Hosts[i] + ": " + err.Error())
				return
			}
			res, err := sshClient.Run("ip addr | grep link/ether")
			// device id is first network interface mac address
			if err != nil {
				dc.client.SendClientError("Could not read mac address from host " + dc.config.Hosts[i] + ": " + err.Error())
				return
			}
			macs := strings.Split(string(res), "\n")
			if len(macs) < 1 || len(macs[0]) < 29 {
				dc.client.SendClientError("Could not determine mac address from host " + dc.config.Hosts[i])
				return
			}

			id := dcPrefix + macs[0][15:32]
			device, ok := devices[id]
			if !ok || device.SshClient == nil {
				device = &model.SshDcDevice{
					DeviceInfo: mgw.DeviceInfo{
						Id:         id,
						Name:       "SSH " + dc.config.Hosts[i],
						DeviceType: dc.config.DeviceTypeId,
					},
					SshClient: sshClient,
				}
			}
			_, err = device.SshClient.Run("echo")
			if err != nil {
				device.State = mgw.Offline
			} else {
				device.State = mgw.Online
			}
			go func() {
				// Set device offline if session closes
				err := device.SshClient.Wait()
				if dc.libConfig.Debug {
					log.Println("DEBUG: SSH to " + dc.config.Hosts[i] + " closed: " + err.Error())
				}
				device.State = mgw.Offline
				err = dc.client.SetDevice(device)
				if err != nil {
					errMsg := "could not send update device : " + err.Error()
					log.Println("ERROR: " + errMsg)
					dc.client.SendClientError(errMsg)
					return
				}
			}()
			err = dc.client.SetDevice(device)
			if err != nil {
				errMsg := "could not send update device : " + err.Error()
				log.Println("ERROR: " + errMsg)
				dc.client.SendClientError(errMsg)
				return
			}
			updateIdsMux.Lock()
			updatedIds = append(updatedIds, id)
			updateIdsMux.Unlock()
		}()
	}
	wg.Wait()
	for id, d := range dc.client.GetDevices() {
		if !slices.Contains(updatedIds, id) {
			// This device was not updated, might have been unreachable -> set offline
			d.State = mgw.Offline
			err = dc.client.SetDevice(d)
			if err != nil {
				errMsg := "could not send update device : " + err.Error()
				log.Println("ERROR: " + errMsg)
				dc.client.SendClientError(errMsg)
			}
		}
	}
}
