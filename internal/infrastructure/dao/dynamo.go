package dao

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/sarangkharche/xapien-billing/internal/domain"
)

type OrgDAO interface {
	SaveOrg(org domain.Organisation) error
	GetOrg(id string) (domain.Organisation, error)
}

type DynamoOrgDAO struct {
	Client    *dynamodb.Client
	TableName string
}

func NewDynamoOrgDAO(cfg aws.Config, tableName string) *DynamoOrgDAO {
	return &DynamoOrgDAO{
		Client:    dynamodb.NewFromConfig(cfg),
		TableName: tableName,
	}
}

func (d *DynamoOrgDAO) SaveOrg(org domain.Organisation) error {
	log.Printf("SaveOrg called for: %s, TotalCredits: %d, Plan: %s, TopUpCredits: %d\n",
		org.ID, org.TotalCredits, org.Plan, org.TopUpCredits)

	item, err := attributevalue.MarshalMap(org)
	if err != nil {
		log.Printf("Marshal error: %v\n", err)
		return fmt.Errorf("marshal error: %w", err)
	}

	log.Printf("Marshaled item: %+v\n", item)

	_, err = d.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &d.TableName,
		Item:      item,
	})
	if err != nil {
		log.Printf("PutItem error: %v\n", err)
		return fmt.Errorf("put item error: %w", err)
	}

	log.Printf("SaveOrg succeeded for: %s, TotalCredits: %d\n", org.ID, org.TotalCredits)
	return nil
}

func (d *DynamoOrgDAO) GetOrg(id string) (domain.Organisation, error) {
	resp, err := d.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &d.TableName,
		Key: map[string]types.AttributeValue{
			"org_id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return domain.Organisation{}, fmt.Errorf("get item error: %w", err)
	}
	if resp.Item == nil {
		return domain.Organisation{}, errors.New("organisation not found")
	}

	var org domain.Organisation
	if err := attributevalue.UnmarshalMap(resp.Item, &org); err != nil {
		return domain.Organisation{}, fmt.Errorf("unmarshal error: %w", err)
	}

	return org, nil
}
