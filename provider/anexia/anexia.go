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
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/pkg/errors"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
	"sigs.k8s.io/external-dns/provider"
	"github.com/anexia-it/go-anxcloud/pkg/client"
)

type AnexiaDNSProvider struct {
	provider.BaseProvider
	Client client.Client
}

func NewAnexiaDNSProvider() (AnexiaDNSProvider,error){

}

func (p *AnexiaDNSProvider) Records(ctx context.Context) (endpoints []*endpoint.Endpoint, _ error) {
	zones, err := p.Zones(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "records retrieval failed")
	}

	return p.records(ctx, zones)
}


func (p *AnexiaDNSProvider) ApplyChanges(ctx context.Context, changes *plan.Changes) error {
	zones, err := p.Zones(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to list zones, not applying changes")
	}

	records, ok := ctx.Value(provider.RecordsContextKey).([]*endpoint.Endpoint)
	if !ok {
		var err error
		records, err = p.records(ctx, zones)
		if err != nil {
			log.Errorf("failed to get records while preparing to applying changes: %s", err)
		}
	}

	updateChanges := p.createUpdateChanges(changes.UpdateNew, changes.UpdateOld, records, zones)

	combinedChanges := make([]*route53.Change, 0, len(changes.Delete)+len(changes.Create)+len(updateChanges))
	combinedChanges = append(combinedChanges, p.newChanges(route53.ChangeActionCreate, changes.Create, records, zones)...)
	combinedChanges = append(combinedChanges, p.newChanges(route53.ChangeActionDelete, changes.Delete, records, zones)...)
	combinedChanges = append(combinedChanges, updateChanges...)

	return p.submitChanges(ctx, combinedChanges, zones)
}

