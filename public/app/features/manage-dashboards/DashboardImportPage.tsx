import React, { FormEvent, PureComponent } from 'react';
import { connect, MapDispatchToProps, MapStateToProps } from 'react-redux';
import { css } from 'emotion';
import { AppEvents, NavModel } from '@grafana/data';
import { Forms, stylesFactory } from '@grafana/ui';
import Page from 'app/core/components/Page/Page';
import { ImportDashboardOverview } from './components/ImportDashboardOverview';
import { DashboardFileUpload } from './components/DashboardFileUpload';
import { fetchGcomDashboard, importDashboardJson } from './state/actions';
import appEvents from 'app/core/app_events';
import { getNavModel } from 'app/core/selectors/navModel';
import { StoreState } from 'app/types';

interface OwnProps {}

interface ConnectedProps {
  navModel: NavModel;
  isLoaded: boolean;
}

interface DispatchProps {
  fetchGcomDashboard: typeof fetchGcomDashboard;
  importDashboardJson: typeof importDashboardJson;
}

type Props = OwnProps & ConnectedProps & DispatchProps;

class DashboardImportUnConnected extends PureComponent<Props> {
  onFileUpload = (event: FormEvent<HTMLInputElement>) => {
    const { importDashboardJson } = this.props;
    const file = event.currentTarget.files[0];

    const reader = new FileReader();
    const readerOnLoad = () => {
      return (e: any) => {
        let dashboard: any;
        try {
          dashboard = JSON.parse(e.target.result);
        } catch (error) {
          appEvents.emit(AppEvents.alertError, ['Import failed', 'JSON -> JS Serialization failed: ' + error.message]);
          return;
        }
        importDashboardJson(dashboard);
      };
    };
    reader.onload = readerOnLoad();
    reader.readAsText(file);
  };

  validateDashboardJson = (json: string) => {
    try {
      JSON.parse(json);
      return true;
    } catch (error) {
      return 'Not valid JSON';
    }
  };

  validateGcomDashboard = (gcomDashboard: string) => {
    // From DashboardImportCtrl
    const match = /(^\d+$)|dashboards\/(\d+)/.exec(gcomDashboard);

    return match && (match[1] || match[2]) ? true : 'Could not find a valid Grafana.com id';
  };

  getDashboardFromJson = (formData: { dashboardJson: string }) => {
    this.props.importDashboardJson(JSON.parse(formData.dashboardJson));
  };

  getGcomDashboard = (formData: { gcomDashboard: string }) => {
    let dashboardId;
    const match = /(^\d+$)|dashboards\/(\d+)/.exec(formData.gcomDashboard);
    if (match && match[1]) {
      dashboardId = match[1];
    } else if (match && match[2]) {
      dashboardId = match[2];
    }

    if (dashboardId) {
      this.props.fetchGcomDashboard(dashboardId);
    }
  };

  renderImportForm() {
    const styles = importStyles();

    return (
      <>
        <div className={styles.option}>
          <DashboardFileUpload onFileUpload={this.onFileUpload} />
        </div>
        <div className={styles.option}>
          <Forms.Legend>Import via grafana.com</Forms.Legend>
          <Forms.Form onSubmit={this.getGcomDashboard} defaultValues={{ gcomDashboard: '' }}>
            {({ register, errors }) => (
              <Forms.Field
                invalid={!!errors.gcomDashboard}
                error={errors.gcomDashboard && errors.gcomDashboard.message}
              >
                <Forms.Input
                  size="md"
                  name="gcomDashboard"
                  placeholder="Grafana.com dashboard url or id"
                  type="text"
                  ref={register({
                    required: 'A Grafana dashboard url or id is required',
                    validate: v => this.validateGcomDashboard(v),
                  })}
                  addonAfter={<Forms.Button type="submit">Load</Forms.Button>}
                />
              </Forms.Field>
            )}
          </Forms.Form>
        </div>
        <div className={styles.option}>
          <Forms.Legend>Import via panel json</Forms.Legend>
          <Forms.Form onSubmit={this.getDashboardFromJson} defaultValues={{ dashboardJson: '' }}>
            {({ register, errors }) => (
              <>
                <Forms.Field
                  invalid={!!errors.dashboardJson}
                  error={errors.dashboardJson && errors.dashboardJson.message}
                >
                  <Forms.TextArea
                    name="dashboardJson"
                    ref={register({
                      required: 'Need a dashboard json model',
                      validate: v => this.validateDashboardJson(v),
                    })}
                    rows={10}
                  />
                </Forms.Field>
                <Forms.Button type="submit">Load</Forms.Button>
              </>
            )}
          </Forms.Form>
        </div>
      </>
    );
  }

  render() {
    const { isLoaded, navModel } = this.props;
    return (
      <Page navModel={navModel}>
        <Page.Contents>{isLoaded ? <ImportDashboardOverview /> : this.renderImportForm()}</Page.Contents>
      </Page>
    );
  }
}

const mapStateToProps: MapStateToProps<ConnectedProps, OwnProps, StoreState> = (state: StoreState) => ({
  navModel: getNavModel(state.navIndex, 'import', null, true),
  isLoaded: state.importDashboard.isLoaded,
});

const mapDispatchToProps: MapDispatchToProps<DispatchProps, Props> = {
  fetchGcomDashboard,
  importDashboardJson,
};

export const DashboardImportPage = connect(mapStateToProps, mapDispatchToProps)(DashboardImportUnConnected);
export default DashboardImportPage;
DashboardImportPage.displayName = 'DashboardImport';

const importStyles = stylesFactory(() => {
  return {
    option: css`
      margin-bottom: 32px;
    `,
  };
});
