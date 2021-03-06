package edgegroup

import (
	"github.com/boltdb/bolt"
	"github.com/cloudogu/portainer-ce/api"
	"github.com/cloudogu/portainer-ce/api/bolt/internal"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "edgegroups"
)

// Service represents a service for managing Edge group data.
type Service struct {
	db *bolt.DB
}

// NewService creates a new instance of a service.
func NewService(db *bolt.DB) (*Service, error) {
	err := internal.CreateBucket(db, BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		db: db,
	}, nil
}

// EdgeGroups return an array containing all the Edge groups.
func (service *Service) EdgeGroups() ([]portainer.EdgeGroup, error) {
	var groups = make([]portainer.EdgeGroup, 0)

	err := service.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var group portainer.EdgeGroup
			err := internal.UnmarshalObjectWithJsoniter(v, &group)
			if err != nil {
				return err
			}
			groups = append(groups, group)
		}

		return nil
	})

	return groups, err
}

// EdgeGroup returns an Edge group by ID.
func (service *Service) EdgeGroup(ID portainer.EdgeGroupID) (*portainer.EdgeGroup, error) {
	var group portainer.EdgeGroup
	identifier := internal.Itob(int(ID))

	err := internal.GetObject(service.db, BucketName, identifier, &group)
	if err != nil {
		return nil, err
	}

	return &group, nil
}

// UpdateEdgeGroup updates an Edge group.
func (service *Service) UpdateEdgeGroup(ID portainer.EdgeGroupID, group *portainer.EdgeGroup) error {
	identifier := internal.Itob(int(ID))
	return internal.UpdateObject(service.db, BucketName, identifier, group)
}

// DeleteEdgeGroup deletes an Edge group.
func (service *Service) DeleteEdgeGroup(ID portainer.EdgeGroupID) error {
	identifier := internal.Itob(int(ID))
	return internal.DeleteObject(service.db, BucketName, identifier)
}

// CreateEdgeGroup assign an ID to a new Edge group and saves it.
func (service *Service) CreateEdgeGroup(group *portainer.EdgeGroup) error {
	return service.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		id, _ := bucket.NextSequence()
		group.ID = portainer.EdgeGroupID(id)

		data, err := internal.MarshalObject(group)
		if err != nil {
			return err
		}

		return bucket.Put(internal.Itob(int(group.ID)), data)
	})
}
