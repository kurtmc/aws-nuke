package resources

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/rebuy-de/aws-nuke/pkg/types"
)

type SageMakerApp struct {
	svc             *sagemaker.SageMaker
	domainId        *string
	appName         *string
	appType         *string
	userProfileName *string
}

func init() {
	register("SageMakerApp", ListSageMakerApps)
}

func ListSageMakerApps(sess *session.Session) ([]Resource, error) {
	svc := sagemaker.New(sess)
	resources := []Resource{}

	params := &sagemaker.ListAppsInput{
		MaxResults: aws.Int64(30),
	}

	for {
		resp, err := svc.ListApps(params)
		if err != nil {
			return nil, err
		}

		for _, app := range resp.Apps {
			resources = append(resources, &SageMakerApp{
				svc:             svc,
				domainId:        app.DomainId,
				appName:         app.AppName,
				appType:         app.AppType,
				userProfileName: app.UserProfileName,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

func (f *SageMakerApp) Remove() error {
	_, err := f.svc.DeleteApp(&sagemaker.DeleteAppInput{
		DomainId:        f.domainId,
		AppName:         f.appName,
		AppType:         f.appType,
		UserProfileName: f.userProfileName,
	})

	return err
}

func (f *SageMakerApp) String() string {
	return *f.appName
}

func (i *SageMakerApp) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("DomainId", i.domainId).
		Set("AppName", i.appName).
		Set("AppType", i.appType).
		Set("UserProfileName", i.userProfileName)
	return properties
}