package cloudwatch

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/tsdb"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLogActions(t *testing.T) {
	Convey("TestLogActions", t, func() {
		executor := &CloudWatchExecutor{
			logsClient: &FakeLogsClient{},
		}

		Convey("handleDescribeLogGroups is valid", func() {
			Convey("when logGroupNamePrefix is empty", func() {
				params := simplejson.NewFromAny(map[string]interface{}{
					"limit": 50,
				})

				frame, err := executor.handleDescribeLogGroups(params)

				expectedField := data.NewField("logGroupName", nil, []*string{aws.String("group_a"), aws.String("group_b"), aws.String("group_c")})
				expectedFrame := data.NewFrame("logGroups", expectedField)
				So(err, ShouldEqual, nil)
				So(frame, ShouldResemble, expectedFrame)
			})

			Convey("when logGroupNamePrefix is not empty", func() {
				params := simplejson.NewFromAny(map[string]interface{}{
					"logGroupNamePrefix": "g",
				})

				frame, err := executor.handleDescribeLogGroups(params)

				expectedField := data.NewField("logGroupName", nil, []*string{aws.String("group_a"), aws.String("group_b"), aws.String("group_c")})
				expectedFrame := data.NewFrame("logGroups", expectedField)
				So(err, ShouldEqual, nil)
				So(frame, ShouldResemble, expectedFrame)
			})
		})

		Convey("handleGetLogGroupFields is valid", func() {

			Convey("when logGroupNamePrefix is not empty", func() {
				params := simplejson.NewFromAny(map[string]interface{}{
					"logGroupName": "group_a",
					"limit":        50,
				})

				frame, err := executor.handleGetLogGroupFields(params, "A")

				expectedNameField := data.NewField("name", nil, []*string{aws.String("field_a"), aws.String("field_b"), aws.String("field_c")})
				expectedPercentField := data.NewField("percent", nil, []*int64{aws.Int64(100), aws.Int64(30), aws.Int64(55)})
				expectedFrame := data.NewFrame("A", expectedNameField, expectedPercentField)
				expectedFrame.RefID = "A"

				So(err, ShouldEqual, nil)
				So(frame, ShouldResemble, expectedFrame)
			})
		})

		Convey("executeStartQuery", func() {

			Convey("should throw when timerange end precedes timerange start", func() {
				timeRange := &tsdb.TimeRange{
					From: "1584873443000",
					To:   "1584700643000",
				}

				params := simplejson.NewFromAny(map[string]interface{}{
					"region":      "default",
					"limit":       50,
					"queryString": "fields @message",
				})

				response, err := executor.executeStartQuery(params, timeRange)

				So(response, ShouldEqual, nil)
				So(err, ShouldResemble, fmt.Errorf("Invalid time range: Start time must be before end time"))
			})
		})

		Convey("handleStartQuery", func() {
			Convey("should return query ID", func() {
				timeRange := &tsdb.TimeRange{
					From: "1584700643000",
					To:   "1584873443000",
				}

				params := simplejson.NewFromAny(map[string]interface{}{
					"region":      "default",
					"limit":       50,
					"queryString": "fields @message",
				})

				frame, err := executor.handleStartQuery(params, timeRange, "A")

				expectedField := data.NewField("queryId", nil, []string{"abcd-efgh-ijkl-mnop"})
				expectedFrame := data.NewFrame("A", expectedField)
				expectedFrame.RefID = "A"
				So(err, ShouldEqual, nil)
				So(frame, ShouldResemble, expectedFrame)
			})
		})

		Convey("handleStopQuery", func() {
			Convey("should stop the query", func() {
				params := simplejson.NewFromAny(map[string]interface{}{
					"queryId": "abcd-efgh-ijkl-mnop",
				})

				frame, err := executor.handleStopQuery(params)

				expectedField := data.NewField("success", nil, []bool{true})
				expectedFrame := data.NewFrame("StopQueryResponse", expectedField)

				So(err, ShouldEqual, nil)
				So(frame, ShouldResemble, expectedFrame)
			})
		})

		Convey("handleGetQueryResults", func() {
			Convey("should return query results as data frame", func() {
				params := simplejson.NewFromAny(map[string]interface{}{
					"queryId": "abcd-efgh-ijkl-mnop",
				})

				frame, err := executor.handleGetQueryResults(params, "A")

				expectedTimeField := data.NewField("@timestamp", nil, []*string{
					aws.String("1584700643"), aws.String("1584700843"),
				})
				expectedTimeField.SetConfig(&data.FieldConfig{Title: "Time"})

				expectedFieldB := data.NewField("field_b", nil, []*string{
					aws.String("b_1"), aws.String("b_2"),
				})

				expectedFrame := data.NewFrame("A", expectedTimeField, expectedFieldB)
				expectedFrame.RefID = "A"

				expectedFrame.Meta = &data.QueryResultMeta{
					Custom: map[string]interface{}{
						"Status": "Complete",
						"Statistics": cloudwatchlogs.QueryStatistics{
							BytesScanned:   aws.Float64(512),
							RecordsMatched: aws.Float64(256),
							RecordsScanned: aws.Float64(1024),
						},
					},
				}

				So(err, ShouldEqual, nil)
				So(frame, ShouldResemble, expectedFrame)
			})
		})
	})
}
