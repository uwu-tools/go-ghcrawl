// Copyright 2022 uwu tools Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	"github.com/google/go-github/v42/github"
)

/*
Crawler reference implementations

(From https://github.com/SAP/project-portal-for-innersource/blob/main/docs/CRAWLING.md)

You will have to adapt all of these crawler implementations to your exact setup. However they may give you a good starting points.

* A plain GitHub API call with some post-processing in [jq](https://stedolan.github.io/jq/). This call will query all repos in a GitHub organization with topic `inner-source` and store it in a local file ([oauth token](https://docs.github.com/en/github/authenticating-to-github/keeping-your-account-and-data-secure/creating-a-personal-access-token) with permission `repo` required). This can for example be used to have a quick demo of the portal up and running with your own data.

  ``` sh
  curl -u <username>:<oauth_token> https://api.github.com/search/repositories?q=org:<org>+topic:inner-source | jq '.items' > repos.json
  ```

* GitHub Crawler implementation with Ruby: [spier/innersource-crawler-ruby](https://github.com/spier/innersource-crawler-ruby)
* GitHub Crawler implementation with Python: [zkoppert/innersource-crawler](https://github.com/zkoppert/innersource-crawler)
* [AWS CodeCommit](https://aws.amazon.com/codecommit/) Crawler implementation with Python: [aws-samples/codecommit-crawler-innersource](https://github.com/aws-samples/codecommit-crawler-innersource)

In the following sections we explain the data structure of `repos.json` and how you would populate it with your own crawler. You will also find [Crawler reference implementations](#reference-implementations) that you can use as starting points for your own crawler.
*/

func NewClient() (
	client *github.Client,
	orgs []*github.Organization,
	repos []*github.Repository,
	err error,
) {
	// TODO(http): Consider passing a roundtripper here
	client = github.NewClient(nil)

	// list all organizations for user "justaugustus"
	orgs, _, err = client.Organizations.List(context.Background(), "justaugustus", nil)
	if err != nil {
		return nil, nil, nil, err
	}

	// list public repositories for org "github"
	opt := &github.RepositoryListByOrgOptions{Type: "public"}
	repos, _, err = client.Repositories.ListByOrg(context.Background(), "github", opt)
	if err != nil {
		return nil, nil, nil, err
	}

	return client, orgs, repos, nil
}

// GetRepositoryActivityScore
// (From https://patterns.innersourcecommons.org/p/repository-activity-score)
// Calculate a virtual InnerSource score from stars, watches, commits, and issues.
func GetRepositoryActivityScore(repo *github.Repository) int {
	// initial score is 50 to give active repos with low GitHub KPIs (forks,
	// watchers, stars) a better starting point
	score := 50

	// weighting: forks and watches count most, then stars, add some little score
	// for open issues, too
	// TODO: Does it matter if these values are not populated?
	score += *repo.ForksCount * 5
	score += *repo.SubscribersCount
	score += *repo.StargazersCount / 3
	score += *repo.OpenIssuesCount / 5

	// updated in last 3 months: adds a bonus multiplier between 0..1 to overall
	// score (1 = updated today, 0 = updated more than 100 days ago)
	/*
		let iDaysSinceLastUpdate = (new Date().getTime() - new Date(repo.updated_at).getTime()) / 1000 / 86400;
		iScore = iScore * ((1 + (100 - Math.min(iDaysSinceLastUpdate, 100))) / 100);
	*/

	// evaluate participation stats for the previous 3 months
	/*
		repo._InnerSourceMetadata = repo._InnerSourceMetadata || {};
		if (repo._InnerSourceMetadata.participation) {
				// average commits: adds a bonus multiplier between 0..1 to overall score (1 = >10 commits per week, 0 = less than 3 commits per week)
				let iAverageCommitsPerWeek = repo._InnerSourceMetadata.participation.slice(repo._InnerSourceMetadata.participation.length - 13).reduce((a, b) => a + b) / 13;
				iScore = iScore * ((1 + (Math.min(Math.max(iAverageCommitsPerWeek - 3, 0), 7))) / 7);
		}
	*/

	// boost calculation:
	// all repositories updated in the previous year will receive a boost of
	// maximum 1000 declining by days since last update
	/*
		let iBoost = (1000 - Math.min(iDaysSinceLastUpdate, 365) * 2.74);
	*/

	// gradually scale down boost according to repository creation date to mix
	// with "real" engagement stats
	/*
		let iDaysSinceCreation = (new Date().getTime() - new Date(repo.created_at).getTime()) / 1000 / 86400;
		iBoost *= (365 - Math.min(iDaysSinceCreation, 365)) / 365;
	*/

	// add boost to score
	/*
		iScore += iBoost;
	*/

	// give projects with a meaningful description a static boost of 50
	/*
		iScore += (repo.description?.length > 30 || repo._InnerSourceMetadata.motivation?.length > 30 ? 50 : 0);
	*/

	// give projects with contribution guidelines (CONTRIBUTING.md) file a static
	// boost of 100
	/*
		iScore += (repo._InnerSourceMetadata.guidelines ? 100 : 0);
	*/

	// build in a logarithmic scale for very active projects (open ended but
	// stabilizing around 5000)
	/*
		if (iScore > 3000) {
			iScore = 3000 + Math.log(iScore) * 100;
		}
	*/

	// final score is a rounded value starting from 0 (subtract the initial
	// value)
	/*
		iScore = Math.round(iScore - 50);
	*/

	// add score to metadata on the fly
	/*
		repo._InnerSourceMetadata.score = iScore;
	*/

	return score
}
