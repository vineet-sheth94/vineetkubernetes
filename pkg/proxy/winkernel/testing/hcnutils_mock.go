//go:build windows
// +build windows

/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testing

import (
	"errors"
	"fmt"

	"github.com/Microsoft/hcsshim/hcn"
)

type HcnMock struct {
	epIdCounter       int
	lbIdCounter       int
	endpointMap       map[string]*hcn.HostComputeEndpoint
	loadbalancerMap   map[string]*hcn.HostComputeLoadBalancer
	supportedFeatures hcn.SupportedFeatures
	network           *hcn.HostComputeNetwork
}

func (hcnObj HcnMock) generateEndpointGuid() (endpointId string, endpointName string) {
	hcnObj.epIdCounter++
	endpointId = fmt.Sprintf("EPID-%d", hcnObj.epIdCounter)
	endpointName = fmt.Sprintf("EPName-%d", hcnObj.epIdCounter)
	return
}

func (hcnObj HcnMock) generateLoadbalancerGuid() (loadbalancerId string) {
	hcnObj.lbIdCounter++
	loadbalancerId = fmt.Sprintf("LBID-%d", hcnObj.lbIdCounter)
	return
}

func NewHcnMock(hnsNetwork *hcn.HostComputeNetwork) *HcnMock {
	return &HcnMock{
		epIdCounter:     0,
		lbIdCounter:     0,
		endpointMap:     make(map[string]*hcn.HostComputeEndpoint),
		loadbalancerMap: make(map[string]*hcn.HostComputeLoadBalancer),
		supportedFeatures: hcn.SupportedFeatures{
			Api: hcn.ApiSupport{
				V2: true,
			},
			DSR:           true,
			IPv6DualStack: true,
		},
		network: hnsNetwork,
	}
}

func (hcnObj HcnMock) PopulateQueriedEndpoints(epId, hnsId, ipAddress, mac string, prefixLen uint8) {
	endpoint := &hcn.HostComputeEndpoint{
		Id:                 epId,
		Name:               epId,
		HostComputeNetwork: hnsId,
		IpConfigurations: []hcn.IpConfig{
			{
				IpAddress:    ipAddress,
				PrefixLength: prefixLen,
			},
		},
		MacAddress: mac,
	}

	hcnObj.endpointMap[endpoint.Id] = endpoint
	hcnObj.endpointMap[endpoint.Name] = endpoint
}

func (hcnObj HcnMock) GetNetworkByName(networkName string) (*hcn.HostComputeNetwork, error) {
	return hcnObj.network, nil
}

func (hcnObj HcnMock) GetNetworkByID(networkID string) (*hcn.HostComputeNetwork, error) {
	return hcnObj.network, nil
}

func (hcnObj HcnMock) ListEndpoints() ([]hcn.HostComputeEndpoint, error) {
	var hcnEPList []hcn.HostComputeEndpoint
	for _, ep := range hcnObj.endpointMap {
		hcnEPList = append(hcnEPList, *ep)
	}
	return hcnEPList, nil
}

func (hcnObj HcnMock) ListEndpointsOfNetwork(networkId string) ([]hcn.HostComputeEndpoint, error) {
	var hcnEPList []hcn.HostComputeEndpoint
	for _, ep := range hcnObj.endpointMap {
		if ep.HostComputeNetwork == networkId {
			hcnEPList = append(hcnEPList, *ep)
		}
	}
	return hcnEPList, nil
}

func (hcnObj HcnMock) GetEndpointByID(endpointId string) (*hcn.HostComputeEndpoint, error) {
	if ep, ok := hcnObj.endpointMap[endpointId]; ok {
		return ep, nil
	}
	epNotFoundError := hcn.EndpointNotFoundError{EndpointID: endpointId}
	return nil, epNotFoundError
}

func (hcnObj HcnMock) GetEndpointByName(endpointName string) (*hcn.HostComputeEndpoint, error) {
	if ep, ok := hcnObj.endpointMap[endpointName]; ok {
		return ep, nil
	}
	epNotFoundError := hcn.EndpointNotFoundError{EndpointName: endpointName}
	return nil, epNotFoundError
}

func (hcnObj HcnMock) CreateEndpoint(network *hcn.HostComputeNetwork, endpoint *hcn.HostComputeEndpoint) (*hcn.HostComputeEndpoint, error) {
	if _, err := hcnObj.GetNetworkByID(network.Id); err != nil {
		return nil, err
	}
	if _, ok := hcnObj.endpointMap[endpoint.Id]; ok {
		return nil, fmt.Errorf("endpoint id %s already present", endpoint.Id)
	}
	if _, ok := hcnObj.endpointMap[endpoint.Name]; ok {
		return nil, fmt.Errorf("endpoint Name %s already present", endpoint.Name)
	}
	endpoint.Id, endpoint.Name = hcnObj.generateEndpointGuid()
	hcnObj.endpointMap[endpoint.Id] = endpoint
	hcnObj.endpointMap[endpoint.Name] = endpoint
	return endpoint, nil
}

func (hcnObj HcnMock) CreateRemoteEndpoint(network *hcn.HostComputeNetwork, endpoint *hcn.HostComputeEndpoint) (*hcn.HostComputeEndpoint, error) {
	return hcnObj.CreateEndpoint(network, endpoint)
}

func (hcnObj HcnMock) DeleteEndpoint(endpoint *hcn.HostComputeEndpoint) error {
	if _, ok := hcnObj.endpointMap[endpoint.Id]; !ok {
		return hcn.EndpointNotFoundError{EndpointID: endpoint.Id}
	}
	delete(hcnObj.endpointMap, endpoint.Id)
	delete(hcnObj.endpointMap, endpoint.Name)
	return nil
}

func (hcnObj HcnMock) ListLoadBalancers() ([]hcn.HostComputeLoadBalancer, error) {
	var hcnLBList []hcn.HostComputeLoadBalancer
	for _, lb := range hcnObj.loadbalancerMap {
		hcnLBList = append(hcnLBList, *lb)
	}
	return hcnLBList, nil
}

func (hcnObj HcnMock) GetLoadBalancerByID(loadBalancerId string) (*hcn.HostComputeLoadBalancer, error) {
	if lb, ok := hcnObj.loadbalancerMap[loadBalancerId]; ok {
		return lb, nil
	}
	lbNotFoundError := hcn.LoadBalancerNotFoundError{LoadBalancerId: loadBalancerId}
	return nil, lbNotFoundError
}

func (hcnObj HcnMock) CreateLoadBalancer(loadBalancer *hcn.HostComputeLoadBalancer) (*hcn.HostComputeLoadBalancer, error) {
	if _, ok := hcnObj.loadbalancerMap[loadBalancer.Id]; ok {
		return nil, fmt.Errorf("LoadBalancer id %s Already Present", loadBalancer.Id)
	}
	loadBalancer.Id = hcnObj.generateLoadbalancerGuid()
	hcnObj.loadbalancerMap[loadBalancer.Id] = loadBalancer
	return loadBalancer, nil
}

func (hcnObj HcnMock) DeleteLoadBalancer(loadBalancer *hcn.HostComputeLoadBalancer) error {
	if _, ok := hcnObj.loadbalancerMap[loadBalancer.Id]; !ok {
		return hcn.LoadBalancerNotFoundError{LoadBalancerId: loadBalancer.Id}
	}
	delete(hcnObj.loadbalancerMap, loadBalancer.Id)
	return nil
}

func (hcnObj HcnMock) GetSupportedFeatures() hcn.SupportedFeatures {
	return hcnObj.supportedFeatures
}

func (hcnObj HcnMock) Ipv6DualStackSupported() error {
	if hcnObj.supportedFeatures.IPv6DualStack {
		return nil
	}
	return errors.New("IPV6 DualStack Not Supported")
}

func (hcnObj HcnMock) DsrSupported() error {
	if hcnObj.supportedFeatures.DSR {
		return nil
	}
	return errors.New("DSR Not Supported")
}

func (hcnObj HcnMock) DeleteAllHnsLoadBalancerPolicy() {
	for k := range hcnObj.loadbalancerMap {
		delete(hcnObj.loadbalancerMap, k)
	}
}
