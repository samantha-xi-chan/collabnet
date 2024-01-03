package service_link

import (
	"collab-net-v2/api"
	"collab-net-v2/link/api_link"
	"collab-net-v2/link/repo_link"
	"context"
	"github.com/pkg/errors"
)

func GetValidLinkFromHostName(hostName string) (links []repo_link.Link, ee error) {
	links, e := repo_link.GetLinkCtl().GetItemsByKeyValue("host_name", hostName)
	if e != nil {
		return nil, errors.Wrap(e, "repo_link.GetLinkCtl().GetItemByKeyValue : ")
	}

	return links, nil
}

func GetLinkItemFromId(ctx context.Context, linkId string) (link repo_link.Link, ee error) {
	item, e := repo_link.GetLinkCtl().GetItemById(ctx, linkId)
	if e != nil {
		return repo_link.Link{}, errors.Wrap(e, "repo_link.GetLinkCtl().GetItemById : ")
	}

	return item, nil
}

func GetFirstPartyNodeLinks(ctx context.Context) (links []repo_link.Link, ee error) {
	var arr []repo_link.QueryKeyValue
	arr = append(arr, repo_link.QueryKeyValue{
		"first_party",
		api.TRUE,
	})
	arr = append(arr, repo_link.QueryKeyValue{
		"online",
		api.TRUE,
	})
	links, e := repo_link.GetLinkCtl().GetItemByKeyValueArr(arr)
	if e != nil {
		return nil, errors.Wrap(e, "repo_link.GetLinkCtl().GetItemByKeyValue : ")
	}

	return links, nil
}

func GetNonFirstPartyNodeLinks(ctx context.Context) (links []repo_link.Link, ee error) {
	var arr []repo_link.QueryKeyValue
	arr = append(arr, repo_link.QueryKeyValue{
		"first_party",
		api.FALSE,
	})
	arr = append(arr, repo_link.QueryKeyValue{
		"online",
		api.TRUE,
	})
	links, e := repo_link.GetLinkCtl().GetItemByKeyValueArr(arr)
	if e != nil {
		return nil, errors.Wrap(e, "repo_link.GetLinkCtl().GetItemByKeyValue : ")
	}

	return links, nil
}

func GetNodes(ctx context.Context, firstParty int, online int) ([]api_link.ApiNode, error) {
	var arr []repo_link.QueryKeyValue
	arr = append(arr, repo_link.QueryKeyValue{
		"first_party",
		interface{}(firstParty), //api.FALSE,
	})
	arr = append(arr, repo_link.QueryKeyValue{
		"online",
		interface{}(online), //api.FALSE,
	})

	links, e := repo_link.GetLinkCtl().GetItemByKeyValueArr(arr)
	if e != nil {
		return nil, errors.Wrap(e, "repo_link.GetLinkCtl().GetItemByKeyValue : ")
	}

	var nodes []api_link.ApiNode
	for _, v := range links {
		nodes = append(nodes, api_link.ApiNode{
			LinkId:     v.Id,
			HostName:   v.HostName,
			FirstParty: v.FirstParty,
			From:       v.From,
			Online:     v.Online,
		})
	}

	return nodes, nil
}
