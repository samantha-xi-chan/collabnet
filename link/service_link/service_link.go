package service_link

import (
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
	links, e := repo_link.GetLinkCtl().GetItemsByKeyValue("first_party", 1)
	if e != nil {
		return nil, errors.Wrap(e, "repo_link.GetLinkCtl().GetItemByKeyValue : ")
	}

	return links, nil
}
