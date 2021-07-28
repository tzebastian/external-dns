package anexia

/*
Copyright 2017 The Kubernetes Authors.

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

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
	"sigs.k8s.io/external-dns/provider"
	"github.com/anexia-it/go-anxcloud/pkg/client"
	"github.com/anexia-it/go-anxcloud/pkg/clouddns/zone"
)

type AnexiaDNSProvider struct {
	provider.BaseProvider
	Client client.Client
}

type AnexiaDNSChange struct {
	Action string
	ResourceRecordSet zone.Record
}

const defaultTTL int = 3600

func NewAnexiaDNSProvider() (*AnexiaDNSProvider,error){
	client,err:=client.New(client.AuthFromEnv(false))
	if err != nil {
		return nil,fmt.Errorf("unable to create client %v", err)
	}
	prov:=&AnexiaDNSProvider{
		Client: client,
	}

	return prov,nil
}

func (p *AnexiaDNSProvider) Records(ctx context.Context) (endpoints []*endpoint.Endpoint,_ error) {
	zones, err := zone.NewAPI(p.Client).List(ctx)

	if err != nil {
		return nil, errors.Wrap(err, "records retrieval failed")
	}

	for _, _zone:=range zones{
		records,err:=zone.NewAPI(p.Client).ListRecords(ctx,_zone.ZoneName)

		if err != nil {
			return nil, errors.Wrap(err, "error retrieving records")
		}

		for _,_record:=range records{
			endpoints=append(endpoints, endpoint.NewEndpointWithTTL(_record.Name,_record.Type,endpoint.TTL(int64(*_record.TTL)),_record.RData))
		}
	}

	return endpoints,nil
}

func (p *AnexiaDNSProvider) ApplyChanges(ctx context.Context, changes *plan.Changes) error {
	combinedChanges := make([]*AnexiaDNSChange, 0, len(changes.Create)+len(changes.UpdateNew)+len(changes.Delete))

	combinedChanges = append(combinedChanges, newAnexiaDNSChanges("CREATE", changes.Create)...)
	combinedChanges = append(combinedChanges, newAnexiaDNSChanges("UPDATE", changes.UpdateNew)...)
	combinedChanges = append(combinedChanges, newAnexiaDNSChanges("DELETE", changes.Delete)...)

	return p.submitChanges(ctx, combinedChanges)
}

func newAnexiaDNSChanges(action string, endpoints []*endpoint.Endpoint) []*AnexiaDNSChange {
	changes := make([]*AnexiaDNSChange, 0, len(endpoints))
	ttl := defaultTTL

	for _, ep := range endpoints {
		if ep.RecordTTL.IsConfigured() {
			ttl = int(ep.RecordTTL)
		}

		change := &AnexiaDNSChange{
			Action: action,
			ResourceRecordSet: zone.Record{
				Name:       ep.DNSName,
				RData:      ep.Targets[0],
				TTL:        &ttl,
				Type:       ep.RecordType,
			},
		}

		changes = append(changes, change)
	}
	return changes
}

func (p *AnexiaDNSProvider) submitChanges(ctx context.Context, changes []*AnexiaDNSChange) error {
	if len(changes) == 0{
		return nil
	}

	zones, err := zone.NewAPI(p.Client).List(ctx)
	if err != nil {
		return err
	}

	separatedChanges := separateChangesByZones(zones, changes)

	for zoneName, changes := range separatedChanges {
		zone.NewAPI(p.Client)
	}
}

func separateChangesByZones(zones []zone.Zone, changes []*AnexiaDNSChange) map[string][]*AnexiaDNSChange {
	change := make(map[string][]*AnexiaDNSChange)
	zoneNameID := provider.ZoneIDName{}

	for _, z := range zones {
		zoneNameID.Add(z.Name, z.Name)
		change[z.Name] = []*AnexiaDNSChange{}
	}

	for _, c := range changes {
		zone, _ := zoneNameID.FindZone(c.ResourceRecordSet.Name)
	}
}

