package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// DynamoDB implementation for DB interface.
type DynamoDB struct {
	config  *aws.Config
	session *session.Session
	db      *dynamodb.DynamoDB
}

// NewDynamoDB creates a new AWS DynamoDB instance.
// region = Optional region setting. Will overwrite AWS_REGION env var if available.
// endpoint = Optional endpoint URL setting. Useful for specifying local/development service URL.
func NewDynamoDB(region string, endpoint string) *DynamoDB {
	db := DynamoDB{}

	// carefully crafting config elements not to mess with the defaults
	if region != "" || endpoint != "" {
		db.config = aws.NewConfig()

		if region != "" {
			db.config.WithRegion(region)
		}
		if endpoint != "" {
			db.config.WithEndpoint(endpoint)
		}

		db.session = session.New(db.config)
	} else {
		db.session = session.New()
	}

	db.db = dynamodb.New(db.session)

	return &db
}

// ListTables lists all tables.
func (db *DynamoDB) ListTables() ([]*string, error) {
	res, err := db.db.ListTables(&dynamodb.ListTablesInput{Limit: aws.Int64(100)})
	if err != nil {
		return nil, err
	}

	return res.TableNames, nil
}

// Seed creates and populates the database, overwriting existing data if specified.
func (db *DynamoDB) Seed(overwrite bool) error {
	params := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{ // Required
			{ // Required
				AttributeName: aws.String("KeySchemaAttributeName"), // Required
				AttributeType: aws.String("ScalarAttributeType"),    // Required
			},
			// More values...
		},
		KeySchema: []*dynamodb.KeySchemaElement{ // Required
			{ // Required
				AttributeName: aws.String("KeySchemaAttributeName"), // Required
				KeyType:       aws.String("KeyType"),                // Required
			},
			// More values...
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{ // Required
			ReadCapacityUnits:  aws.Int64(1), // Required
			WriteCapacityUnits: aws.Int64(1), // Required
		},
		TableName: aws.String("TableName"), // Required
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{ // Required
				IndexName: aws.String("IndexName"), // Required
				KeySchema: []*dynamodb.KeySchemaElement{ // Required
					{ // Required
						AttributeName: aws.String("KeySchemaAttributeName"), // Required
						KeyType:       aws.String("KeyType"),                // Required
					},
					// More values...
				},
				Projection: &dynamodb.Projection{ // Required
					NonKeyAttributes: []*string{
						aws.String("NonKeyAttributeName"), // Required
						// More values...
					},
					ProjectionType: aws.String("ProjectionType"),
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{ // Required
					ReadCapacityUnits:  aws.Int64(1), // Required
					WriteCapacityUnits: aws.Int64(1), // Required
				},
			},
			// More values...
		},
		LocalSecondaryIndexes: []*dynamodb.LocalSecondaryIndex{
			{ // Required
				IndexName: aws.String("IndexName"), // Required
				KeySchema: []*dynamodb.KeySchemaElement{ // Required
					{ // Required
						AttributeName: aws.String("KeySchemaAttributeName"), // Required
						KeyType:       aws.String("KeyType"),                // Required
					},
					// More values...
				},
				Projection: &dynamodb.Projection{ // Required
					NonKeyAttributes: []*string{
						aws.String("NonKeyAttributeName"), // Required
						// More values...
					},
					ProjectionType: aws.String("ProjectionType"),
				},
			},
			// More values...
		},
		StreamSpecification: &dynamodb.StreamSpecification{
			StreamEnabled:  aws.Bool(true),
			StreamViewType: aws.String("StreamViewType"),
		},
	}

	_, err := db.db.CreateTable(params)
	if err != nil {
		return err
	}

	// wait till all tables are read for use
	// db.db.WaitUntilTableExists

	return nil
}
