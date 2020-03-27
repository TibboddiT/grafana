package cloudwatch

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/tsdb"
)

func (e *CloudWatchExecutor) executeLogActions(queryContext *tsdb.TsdbQuery) (*tsdb.Response, error) {
	response := &tsdb.Response{
		Results: make(map[string]*tsdb.QueryResult),
	}

	for _, query := range queryContext.Queries {
		dataframe, err := e.executeLogAction(queryContext, query)
		if dataframe == nil {
			return nil, err
		}

		dataframeEnc, err := data.MarshalArrow(dataframe)

		response.Results[query.RefId] = &tsdb.QueryResult{RefId: query.RefId, Dataframes: [][]byte{dataframeEnc}}
	}

	return response, nil
}

func (e *CloudWatchExecutor) executeLogAction(queryContext *tsdb.TsdbQuery, query *tsdb.Query) (*data.Frame, error) {
	parameters := query.Model
	subType := query.Model.Get("subtype").MustString()

	var data *data.Frame = nil
	var err error = nil

	switch subType {
	case "DescribeLogGroups":
		data, err = e.handleDescribeLogGroups(parameters)
	case "GetLogGroupFields":
		data, err = e.handleGetLogGroupFields(parameters, query.RefId)
	case "StartQuery":
		data, err = e.handleStartQuery(parameters, queryContext.TimeRange, query.RefId)
	case "StopQuery":
		data, err = e.handleStopQuery(parameters)
	case "GetQueryResults":
		data, err = e.handleGetQueryResults(parameters, query.RefId)
	}

	if data == nil {
		return nil, err
	}

	return data, nil
}

func (e *CloudWatchExecutor) handleDescribeLogGroups(parameters *simplejson.Json) (*data.Frame, error) {
	logGroupNamePrefix := parameters.Get("logGroupNamePrefix").MustString("")
	var response *cloudwatchlogs.DescribeLogGroupsOutput = nil
	var err error = nil

	if len(logGroupNamePrefix) < 1 {
		response, err = e.logsClient.DescribeLogGroups(&cloudwatchlogs.DescribeLogGroupsInput{
			Limit: aws.Int64(parameters.Get("limit").MustInt64(50)),
		})
	} else {
		response, err = e.logsClient.DescribeLogGroups(&cloudwatchlogs.DescribeLogGroupsInput{
			Limit:              aws.Int64(parameters.Get("limit").MustInt64(50)),
			LogGroupNamePrefix: aws.String(logGroupNamePrefix),
		})
	}

	if err != nil || response == nil {
		return nil, err.(awserr.Error)
	}

	logGroupNames := make([]*string, 0)
	for _, logGroup := range response.LogGroups {
		logGroupNames = append(logGroupNames, logGroup.LogGroupName)
	}

	groupNamesField := data.NewField("logGroupName", nil, logGroupNames)
	frame := data.NewFrame("logGroups", groupNamesField)

	return frame, nil
}

func (e *CloudWatchExecutor) executeStartQuery(parameters *simplejson.Json, timeRange *tsdb.TimeRange) (*cloudwatchlogs.StartQueryOutput, error) {
	startTime, err := timeRange.ParseFrom()
	if err != nil {
		return nil, err
	}

	endTime, err := timeRange.ParseTo()
	if err != nil {
		return nil, err
	}

	if !startTime.Before(endTime) {
		return nil, fmt.Errorf("Invalid time range: Start time must be before end time")
	}

	requestQueries := make(map[string][]*logsQuery)
	region := parameters.Get("region").MustString("")

	if _, exist := requestQueries[region]; !exist {
		requestQueries[region] = make([]*logsQuery, 0)
	}

	startQueryInput := &cloudwatchlogs.StartQueryInput{
		StartTime:     aws.Int64(startTime.Unix()),
		EndTime:       aws.Int64(endTime.Unix()),
		Limit:         aws.Int64(parameters.Get("limit").MustInt64(1000)),
		LogGroupNames: aws.StringSlice(parameters.Get("logGroupNames").MustStringArray()),
		QueryString:   aws.String("fields @timestamp | " + parameters.Get("queryString").MustString("")),
	}
	return e.logsClient.StartQuery(startQueryInput)
}

func (e *CloudWatchExecutor) handleStartQuery(parameters *simplejson.Json, timeRange *tsdb.TimeRange, refID string) (*data.Frame, error) {
	startQueryResponse, err := e.executeStartQuery(parameters, timeRange)
	if err != nil {
		return nil, fmt.Errorf(err.(awserr.Error).Message())
	}

	dataFrame := data.NewFrame(refID, data.NewField("queryId", nil, []string{*startQueryResponse.QueryId}))
	dataFrame.RefID = refID

	return dataFrame, nil
}

func (e *CloudWatchExecutor) executeStopQuery(parameters *simplejson.Json) (*cloudwatchlogs.StopQueryOutput, error) {
	queryInput := &cloudwatchlogs.StopQueryInput{
		QueryId: aws.String(parameters.Get("queryId").MustString()),
	}

	response, err := e.logsClient.StopQuery(queryInput)
	if err != nil {
		awsErr, _ := err.(awserr.Error)
		if awsErr.Code() == "InvalidParameterException" {
			response = &cloudwatchlogs.StopQueryOutput{Success: aws.Bool(false)}
			err = nil
		} else {
			err = fmt.Errorf(err.(awserr.Error).Message())
		}
	}

	return response, err
}

func (e *CloudWatchExecutor) handleStopQuery(parameters *simplejson.Json) (*data.Frame, error) {
	response, err := e.executeStopQuery(parameters)

	if err != nil {
		return nil, fmt.Errorf(err.(awserr.Error).Message())
	}

	dataFrame := data.NewFrame("StopQueryResponse", data.NewField("success", nil, []bool{*response.Success}))
	return dataFrame, nil
}

func (e *CloudWatchExecutor) executeGetQueryResults(parameters *simplejson.Json) (*cloudwatchlogs.GetQueryResultsOutput, error) {
	queryInput := &cloudwatchlogs.GetQueryResultsInput{
		QueryId: aws.String(parameters.Get("queryId").MustString()),
	}

	return e.logsClient.GetQueryResults(queryInput)
}

func (e *CloudWatchExecutor) handleGetQueryResults(parameters *simplejson.Json, refID string) (*data.Frame, error) {
	getQueryResultsOutput, err := e.executeGetQueryResults(parameters)
	if err != nil {
		return nil, fmt.Errorf(err.(awserr.Error).Message())
	}

	dataFrame := logsResultsToDataframes(getQueryResultsOutput)
	dataFrame.Name = refID
	dataFrame.RefID = refID

	return dataFrame, nil
}

func (e *CloudWatchExecutor) handleGetLogGroupFields(parameters *simplejson.Json, refID string) (*data.Frame, error) {
	queryInput := &cloudwatchlogs.GetLogGroupFieldsInput{
		LogGroupName: aws.String(parameters.Get("logGroupName").MustString()),
		Time:         aws.Int64(parameters.Get("time").MustInt64()),
	}

	getLogGroupFieldsOutput, err := e.logsClient.GetLogGroupFields(queryInput)
	if err != nil {
		return nil, fmt.Errorf(err.(awserr.Error).Message())
	}

	fieldNames := make([]*string, 0)
	fieldPercentages := make([]*int64, 0)

	for _, logGroupField := range getLogGroupFieldsOutput.LogGroupFields {
		fieldNames = append(fieldNames, logGroupField.Name)
		fieldPercentages = append(fieldPercentages, logGroupField.Percent)
	}

	dataFrame := data.NewFrame(
		refID,
		data.NewField("name", nil, fieldNames),
		data.NewField("percent", nil, fieldPercentages),
	)

	dataFrame.RefID = refID

	return dataFrame, nil
}
