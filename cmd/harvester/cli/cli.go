package cli

import (
	"context"
	"fmt"

	"github.com/HewlettPackard/Galadriel/pkg/common"
	"github.com/HewlettPackard/Galadriel/pkg/harvester"
	"github.com/HewlettPackard/Galadriel/pkg/harvester/config"
	"github.com/HewlettPackard/Galadriel/pkg/harvester/spire"
	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"google.golang.org/grpc/codes"
)

const defaultConfPath = "conf/harvester/harvester.conf"

type HarvesterCLI struct {
	logger *common.Logger
}

func NewHarvesterCLI() *HarvesterCLI {
	return &HarvesterCLI{
		logger: common.NewLogger("harvester"),
	}
}

func (c *HarvesterCLI) Run(args []string) int {
	if len(args) != 1 {
		c.logger.Error("Unknown arguments", args)
		return 1
	}

	cfg, err := config.LoadFromDisk(defaultConfPath)

	if err != nil {
		c.logger.Error("Error loading config:", err)
		return 1
	}

	ctx := context.Background()
	if args[0] == "run" {
		harvester.NewHarvesterManager().Start(ctx, *cfg)
	}

	if args[0] == "spire bundle" {
		return c.spireBundle(ctx, cfg)
	}

	if args[0] == "spire federation list" {
		return c.spireFederationList(ctx, cfg)
	}

	if args[0] == "spire federation create" {
		return c.spireFederationCreate(ctx, cfg)
	}

	if args[0] == "spire federation update" {
		return c.spireFederationUpdate(ctx, cfg)
	}

	if args[0] == "spire federation delete" {
		return c.spireFederationDelete(ctx, cfg)
	}

	c.logger.Error("Unknown command:", args[0])
	return 1
}

// TODO: the following methods are only intended for testing and should be replaced by a proper CLI implementation
func (c *HarvesterCLI) spireBundle(ctx context.Context, cfg *config.HarvesterConfig) int {
	server := spire.NewLocalSpireServer(cfg.HarvesterConfigSection.SpireSocketPath)
	bundle, err := server.GetBundle(ctx)
	if err != nil {
		c.logger.Error("Error getting bundle:", err)
		return 1
	}
	bytes, _ := bundle.X509Bundle().Marshal()
	c.logger.Debug(string(bytes))
	return 0
}

func (c *HarvesterCLI) spireFederationList(ctx context.Context, cfg *config.HarvesterConfig) int {
	server := spire.NewLocalSpireServer(cfg.HarvesterConfigSection.SpireSocketPath)
	feds, err := server.ListFederationRelationships(ctx)
	if err != nil {
		c.logger.Error("Error getting federation relationships:", err)
		return 1
	}
	if len(feds) == 0 {
		c.logger.Info("No federation relationships found")
		return 0
	}

	for _, fed := range feds {
		c.logger.Info(fmt.Sprintf("Trust Domain: %s", fed.TrustDomain))
		c.logger.Info(fmt.Sprintf("Bundle Endpoint Profile: %T", fed.BundleEndpointProfile))
		c.logger.Info(fmt.Sprintf("Bundle Endpoint URL: %s", fed.BundleEndpointURL))
		c.logger.Info("------")
	}
	return 0
}

func (c *HarvesterCLI) spireFederationCreate(ctx context.Context, cfg *config.HarvesterConfig) int {
	td := "test.org"
	server := spire.NewLocalSpireServer(cfg.HarvesterConfigSection.SpireSocketPath)
	bundle := spiffebundle.New(spiffeid.RequireTrustDomainFromString(td))
	toCreate := []*spire.FederationRelationship{
		{
			TrustDomain:           spiffeid.RequireTrustDomainFromString(fmt.Sprintf("spiffe://%s", td)),
			BundleEndpointProfile: spire.HTTPSWebBundleEndpointProfile{},
			BundleEndpointURL:     "https://localhost:8080/spire/bundles/spiffe_bundle",
			TrustDomainBundle:     bundle,
		},
	}
	rels, err := server.CreateFederationRelationships(ctx, toCreate)
	if err != nil {
		c.logger.Error("Error creating federation relationship:", err)
		return 1
	}

	if len(rels) == 0 {
		c.logger.Error("No federation relationships were created")
		return 1
	}

	for _, rel := range rels {
		c.logger.Info(fmt.Sprintf("Code %s: %s", rel.Status.Code.String(), rel.Status.Message))
		if rel.Status.Code == codes.OK {
			c.logger.Info(fmt.Sprintf("Trust Domain: %s", rel.FederationRelationship.TrustDomain))
			c.logger.Info(fmt.Sprintf("Bundle Endpoint Profile: %T", rel.FederationRelationship.BundleEndpointProfile))
			c.logger.Info(fmt.Sprintf("Bundle Endpoint URL: %s", rel.FederationRelationship.BundleEndpointURL))
			c.logger.Info("------")
		}
	}
	return 0
}

func (c *HarvesterCLI) spireFederationUpdate(ctx context.Context, cfg *config.HarvesterConfig) int {
	td := "test.org"
	server := spire.NewLocalSpireServer(cfg.HarvesterConfigSection.SpireSocketPath)
	bundle := spiffebundle.New(spiffeid.RequireTrustDomainFromString(td))
	toUpdate := []*spire.FederationRelationship{
		{
			TrustDomain:           spiffeid.RequireTrustDomainFromString(fmt.Sprintf("spiffe://%s", td)),
			BundleEndpointProfile: spire.HTTPSWebBundleEndpointProfile{},
			BundleEndpointURL:     "https://localhost:8080/spire/bundles/updated",
			TrustDomainBundle:     bundle,
		},
	}
	rels, err := server.UpdateFederationRelationships(ctx, toUpdate)
	if err != nil {
		c.logger.Error("Error updating federation relationship:", err)
		return 1
	}

	if len(rels) == 0 {
		c.logger.Error("No federation relationships were updated")
		return 1
	}

	for _, rel := range rels {
		c.logger.Info(fmt.Sprintf("Code %s: %s", rel.Status.Code.String(), rel.Status.Message))
		if rel.Status.Code == codes.OK {
			c.logger.Info(fmt.Sprintf("Trust Domain: %s", rel.FederationRelationship.TrustDomain))
			c.logger.Info(fmt.Sprintf("Bundle Endpoint Profile: %T", rel.FederationRelationship.BundleEndpointProfile))
			c.logger.Info(fmt.Sprintf("Bundle Endpoint URL: %s", rel.FederationRelationship.BundleEndpointURL))
			c.logger.Info("------")
		}
	}
	return 0
}

func (c *HarvesterCLI) spireFederationDelete(ctx context.Context, cfg *config.HarvesterConfig) int {
	td := "test.org"
	server := spire.NewLocalSpireServer(cfg.HarvesterConfigSection.SpireSocketPath)
	trustDomain := spiffeid.RequireTrustDomainFromString(td)
	toDelete := []*spiffeid.TrustDomain{&trustDomain}
	rels, err := server.DeleteFederationRelationships(ctx, toDelete)
	if err != nil {
		c.logger.Error("Error deleting federation relationship:", err)
		return 1
	}

	if len(rels) == 0 {
		c.logger.Error("No federation relationships were deleted")
		return 1
	}

	for _, rel := range rels {
		c.logger.Info(fmt.Sprintf("Code %s: %s", rel.Status.Code.String(), rel.Status.Message))
	}
	return 0
}
