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
	"github.com/anexia-it/go-anxcloud/pkg/clouddns/zone"
	uuid "github.com/satori/go.uuid"
)

// Provider defines the interface DNS providers should implement.
type AnexiaProvider interface {
	List(ctx context.Context) ([]zone.Zone, error)
	Get(ctx context.Context, name string) (zone.Zone, error)
	Create(ctx context.Context, create zone.Definition) (zone.Zone, error)
	Update(ctx context.Context, name string, update zone.Definition) (zone.Zone, error)
	Delete(ctx context.Context, name string) error
	Apply(ctx context.Context, name string, changeset zone.ChangeSet) ([]zone.Record, error)
	Import(ctx context.Context, name string, zoneData zone.Import) (zone.Revision, error)
	ListRecords(ctx context.Context, name string) ([]zone.Record, error)
	NewRecord(ctx context.Context, zone string, record zone.RecordRequest) (zone.Zone, error)
	UpdateRecord(ctx context.Context, zone string, id uuid.UUID, record zone.RecordRequest) (zone.Zone, error)
	DeleteRecord(ctx context.Context, zone string, id uuid.UUID) error
}


