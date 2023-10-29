package strikes

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/privateerproj/privateer-sdk/raidengine"
	"github.com/privateerproj/privateer-sdk/utils"
	"github.com/spf13/viper"
)

type Strikes struct {
	Log hclog.Logger
}

type Movement struct {
	Strike string
}

func (a *Strikes) SetLogger(loggerName string) {
	a.Log = raidengine.GetLogger(loggerName, false)
}

func getDBConfig() (string, error) {
	if viper.IsSet("raids.rds.config.host") && viper.IsSet("raids.rds.config.database") {
		return "database_host_placeholder", nil
	}
	return "", errors.New("database url must be set in the config file")
}

func getHostDBInstanceIdentifier() (string, error) {
	if viper.IsSet("raids.rds.config.instance_identifier") {
		return viper.GetString("raids.rds.config.instance_identifier"), nil
	}
	return "", errors.New("database instance identifier must be set in the config file")
}

func getHostRDSRegion() (string, error) {
	if viper.IsSet("raids.rds.config.primary_region") {
		return viper.GetString("raids.rds.config.primary_region"), nil
	}
	return "", errors.New("database instance identifier must be set in the config file")
}

func getAWSConfig() (cfg aws.Config, err error) {
	if viper.IsSet("aws") &&
		viper.IsSet("aws.access_key") &&
		viper.IsSet("aws.secret_key") &&
		viper.IsSet("aws.region") {

		access_key := viper.GetString("aws.access_key")
		secret_key := viper.GetString("aws.secret_key")
		session_key := viper.GetString("aws.session_key")
		region := viper.GetString("aws.region")

		creds := credentials.NewStaticCredentialsProvider(access_key, secret_key, session_key)
		cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithCredentialsProvider(creds), config.WithRegion(region))
	}
	return
}

func connectToDb() (result raidengine.MovementResult) {
	result = raidengine.MovementResult{
		Description: "The database host must be available and accepting connections",
		Function:    utils.CallerPath(0),
	}
	_, err := getDBConfig()
	if err != nil {
		result.Message = err.Error()
		return
	}
	result.Passed = true
	return
}

func checkRDSInstanceMovement(cfg aws.Config) (result raidengine.MovementResult) {
	// check if the instance is available
	result = raidengine.MovementResult{
		Description: "Check if the instance is available/exists",
		Function:    utils.CallerPath(0),
	}

	instanceIdentifier, _ := getHostDBInstanceIdentifier()

	instance, err := getRDSInstanceFromIdentifier(cfg, instanceIdentifier)
	if err != nil {
		// Handle error
		result.Message = err.Error()
		result.Passed = false
		return
	}
	result.Passed = len(instance.DBInstances) > 0
	return
}

func getRDSInstanceFromIdentifier(cfg aws.Config, identifier string) (instance *rds.DescribeDBInstancesOutput, err error) {
	rdsClient := rds.NewFromConfig(cfg)

	input := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(identifier),
	}

	instance, err = rdsClient.DescribeDBInstances(context.TODO(), input)
	return
}
