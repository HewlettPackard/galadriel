package spire

import (
	"context"

	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

type fakeClient struct {
	bundle                           *spiffebundle.Bundle
	federationRelationships          []*FederationRelationship
	getBundleErr                     error
	getFederationRelationshipsErr    error
	createFederationRelationshipsErr error
	updateFederationRelationshipsErr error
	deleteFederationRelationshipsErr error
}

func (c fakeClient) GetBundle(context.Context) (*spiffebundle.Bundle, error) {
	if c.getBundleErr != nil {
		return nil, c.getBundleErr
	}

	return c.bundle, nil
}

func (c fakeClient) ListFederationRelationships(context.Context) ([]*FederationRelationship, error) {
	if c.getFederationRelationshipsErr != nil {
		return nil, c.getFederationRelationshipsErr
	}

	return c.federationRelationships, nil
}

func (c fakeClient) CreateFederationRelationships(context.Context, []*FederationRelationship) ([]*FederationRelationshipResult, error) {
	if c.createFederationRelationshipsErr != nil {
		return nil, c.createFederationRelationshipsErr
	}

	// TODO: add create logic and convert []*FederationRelationship to []*FederationRelationshipResult
	return []*FederationRelationshipResult{}, nil
}

func (c fakeClient) UpdateFederationRelationships(context.Context, []*FederationRelationship) ([]*FederationRelationshipResult, error) {
	if c.updateFederationRelationshipsErr != nil {
		return nil, c.updateFederationRelationshipsErr
	}

	// TODO: add update logic
	return []*FederationRelationshipResult{}, nil
}

func (c fakeClient) DeleteFederationRelationships(context.Context, []*spiffeid.TrustDomain) ([]*FederationRelationshipResult, error) {
	if c.deleteFederationRelationshipsErr != nil {
		return nil, c.deleteFederationRelationshipsErr
	}

	// TODO: add delete logic
	return []*FederationRelationshipResult{}, nil
}
