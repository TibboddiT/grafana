package cloudwatch

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
)

type FakeLogsClient struct {
	cloudwatchlogsiface.CloudWatchLogsAPI
}

func (f FakeLogsClient) DescribeLogGroups(input *cloudwatchlogs.DescribeLogGroupsInput) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
	return &cloudwatchlogs.DescribeLogGroupsOutput{
		LogGroups: []*cloudwatchlogs.LogGroup{
			{
				LogGroupName: aws.String("group_a"),
			},
			{
				LogGroupName: aws.String("group_b"),
			},
			{
				LogGroupName: aws.String("group_c"),
			},
		},
	}, nil
}

func (f FakeLogsClient) GetLogGroupFields(input *cloudwatchlogs.GetLogGroupFieldsInput) (*cloudwatchlogs.GetLogGroupFieldsOutput, error) {
	return &cloudwatchlogs.GetLogGroupFieldsOutput{
		LogGroupFields: []*cloudwatchlogs.LogGroupField{
			{
				Name:    aws.String("field_a"),
				Percent: aws.Int64(100),
			},
			{
				Name:    aws.String("field_b"),
				Percent: aws.Int64(30),
			},
			{
				Name:    aws.String("field_c"),
				Percent: aws.Int64(55),
			},
		},
	}, nil
}

func (f FakeLogsClient) StartQuery(input *cloudwatchlogs.StartQueryInput) (*cloudwatchlogs.StartQueryOutput, error) {
	return &cloudwatchlogs.StartQueryOutput{
		QueryId: aws.String("abcd-efgh-ijkl-mnop"),
	}, nil
}

func (f FakeLogsClient) StopQuery(input *cloudwatchlogs.StopQueryInput) (*cloudwatchlogs.StopQueryOutput, error) {
	return &cloudwatchlogs.StopQueryOutput{
		Success: aws.Bool(true),
	}, nil
}

func (f FakeLogsClient) GetQueryResults(input *cloudwatchlogs.GetQueryResultsInput) (*cloudwatchlogs.GetQueryResultsOutput, error) {
	return &cloudwatchlogs.GetQueryResultsOutput{
		Results: [][]*cloudwatchlogs.ResultField{
			{
				{
					Field: aws.String("@timestamp"),
					Value: aws.String("1584700643"),
				},
				{
					Field: aws.String("field_b"),
					Value: aws.String("b_1"),
				},
				{
					Field: aws.String("@ptr"),
					Value: aws.String("abcdefg"),
				},
			},

			{
				{
					Field: aws.String("@timestamp"),
					Value: aws.String("1584700843"),
				},
				{
					Field: aws.String("field_b"),
					Value: aws.String("b_2"),
				},
				{
					Field: aws.String("@ptr"),
					Value: aws.String("hijklmnop"),
				},
			},
		},

		Statistics: &cloudwatchlogs.QueryStatistics{
			BytesScanned:   aws.Float64(512),
			RecordsMatched: aws.Float64(256),
			RecordsScanned: aws.Float64(1024),
		},

		Status: aws.String("Complete"),
	}, nil
}
