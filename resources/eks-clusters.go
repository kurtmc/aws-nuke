package resources

import (
	"context"

	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	eksTypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/gotidy/ptr"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/settings"
	libsettings "github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EKSClusterResource = "EKSCluster"

func init() {
	registry.Register(&registry.Registration{
		Name:     EKSClusterResource,
		Scope:    nuke.Account,
		Resource: &EKSCluster{},
		Lister:   &EKSClusterLister{},
		Settings: []string{
			"DisableDeletionProtection",
		},
	})
}

type EKSClusterLister struct{}

func (l *EKSClusterLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := eks.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	params := &eks.ListClustersInput{
		MaxResults: aws.Int32(100),
	}

	paginator := eks.NewListClustersPaginator(svc, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, cluster := range resp.Clusters {
			dcResp, err := svc.DescribeCluster(ctx, &eks.DescribeClusterInput{Name: aws.String(cluster)})
			if err != nil {
				return nil, err
			}
			resources = append(resources, &EKSCluster{
				svc:     svc,
				name:    aws.String(cluster),
				cluster: dcResp.Cluster,
			})
		}
	}
	return resources, nil
}

type EKSCluster struct {
	svc     *eks.Client
	name    *string
	cluster *eksTypes.Cluster

	settings *libsettings.Setting
}

func (f *EKSCluster) Remove(ctx context.Context) error {
	if ptr.ToBool(f.cluster.DeletionProtection) && f.settings.GetBool("DisableDeletionProtection") {
		updateClusterConfigInput := &eks.UpdateClusterConfigInput{
			Name:               f.name,
			DeletionProtection: aws.Bool(false),
		}
		if _, err := f.svc.UpdateClusterConfig(ctx, updateClusterConfigInput); err != nil {
			return err
		}
	}
	_, err := f.svc.DeleteCluster(ctx, &eks.DeleteClusterInput{
		Name: f.name,
	})

	return err
}

func (f *EKSCluster) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("CreatedAt", f.cluster.CreatedAt.Format(time.RFC3339))
	for key, value := range f.cluster.Tags {
		properties.SetTag(&key, value)
	}
	return properties
}

func (f *EKSCluster) String() string {
	return *f.name
}

func (f *EKSCluster) Settings(setting *settings.Setting) {
	f.settings = setting
}
