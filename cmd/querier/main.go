package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dgamingfoundation/dwh/common"
	log "github.com/sirupsen/logrus"

	"github.com/graphql-go/graphql"
)

func main() {
	db, err := common.GetDB()
	if err != nil {
		log.Fatalf("failed to establish database connection: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Errorf("failed to close database connection: %v", err)
		}
	}()

	var nftType = graphql.NewObject(
		graphql.ObjectConfig{
			Name: "NFT",
			// we define the name and the fields of our
			// object. In this case, we have one solitary
			// field that is of type string
			Fields: graphql.Fields{
				"id": &graphql.Field{
					Type: graphql.Int,
				},
				"uuid": &graphql.Field{
					Type: graphql.String,
				},
				"owner": &graphql.Field{
					Type: graphql.String,
				},
				"name": &graphql.Field{
					Type: graphql.String,
				},
				"description": &graphql.Field{
					Type: graphql.String,
				},
				"image": &graphql.Field{
					Type: graphql.String,
				},
				"tokenURI": &graphql.Field{
					Type: graphql.String,
				},
				"onSale": &graphql.Field{
					Type: graphql.Boolean,
				},
				"price": &graphql.Field{
					Type: graphql.NewList(graphql.String),
				},
				"sellerBeneficiary": &graphql.Field{
					Type: graphql.String,
				},
			},
		},
	)

	// Schema.
	fields := graphql.Fields{
		// Endpoint for returning a specific NFT by ID.
		"nft": &graphql.Field{
			Type:        nftType,
			Description: "Get an NFT by ID",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				var nft common.NFT
				id, ok := p.Args["id"].(string)
				if !ok {
					return nil, errors.New("invalid uuid string")
				}

				db.Where("id = ?", id).First(&nft)
				if len(nft.TokenID) == 0 {
					return nil, fmt.Errorf("failed to find NFT with ID %s", id)
				}

				return &nft, nil
			},
		},
		// Endpoint for returning all NFTs.
		"list": &graphql.Field{
			Type:        graphql.NewList(nftType),
			Description: "Get a list of all NFTs",
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				var nfts []*common.NFT
				db.Find(&nfts)

				return nfts, nil
			},
		},
	}
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	// Query
	query := `
        {
            list {
				uuid
				name
			}
        }
    `
	params := graphql.Params{Schema: schema, RequestString: query}
	r := graphql.Do(params)
	if len(r.Errors) > 0 {
		log.Fatalf("failed to execute graphql operation, errors: %+v", r.Errors)
	}
	rJSON, _ := json.Marshal(r)
	fmt.Printf("%s \n", rJSON)
}
