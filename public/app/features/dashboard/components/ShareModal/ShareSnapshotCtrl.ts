import angular from 'angular';
import _ from 'lodash';

export class ShareSnapshotCtrl {
  /** @ngInject */
  constructor($scope, $rootScope, $location, backendSrv, $timeout, timeSrv) {
    $scope.snapshot = {
      name: $scope.dashboard.title,
      expires: 0,
      timeoutSeconds: 4,
    };

    $scope.step = 1;

    $scope.expireOptions = [
      { text: '1小时', value: 60 * 60 },
      { text: '1天', value: 60 * 60 * 24 },
      { text: '7天', value: 60 * 60 * 24 * 7 },
      { text: '无', value: 0 },
    ];

    $scope.accessOptions = [
      { text: '任何有链接的人都可以查看', value: 1 },
      { text: '组织用户', value: 2 },
      { text: '公开在网上', value: 3 },
    ];

    $scope.init = () => {
      backendSrv.get('/api/snapshot/shared-options').then(options => {
        $scope.sharingButtonText = options['externalSnapshotName'];
        $scope.externalEnabled = options['externalEnabled'];
      });
    };

    $scope.apiUrl = '/api/snapshots';

    $scope.createSnapshot = external => {
      $scope.dashboard.snapshot = {
        timestamp: new Date(),
      };

      if (!external) {
        $scope.dashboard.snapshot.originalUrl = $location.absUrl();
      }

      $scope.loading = true;
      $scope.snapshot.external = external;
      $scope.dashboard.startRefresh();

      $timeout(() => {
        $scope.saveSnapshot(external);
      }, $scope.snapshot.timeoutSeconds * 1000);
    };

    $scope.saveSnapshot = external => {
      const dash = $scope.dashboard.getSaveModelClone();
      $scope.scrubDashboard(dash);

      const cmdData = {
        dashboard: dash,
        name: dash.title,
        expires: $scope.snapshot.expires,
        external: external,
      };

      backendSrv.post($scope.apiUrl, cmdData).then(
        results => {
          $scope.loading = false;
          $scope.deleteUrl = results.deleteUrl;
          $scope.snapshotUrl = results.url;
          $scope.step = 2;
        },
        () => {
          $scope.loading = false;
        }
      );
    };

    $scope.getSnapshotUrl = () => {
      return $scope.snapshotUrl;
    };

    $scope.scrubDashboard = dash => {
      // change title
      dash.title = $scope.snapshot.name;

      // make relative times absolute
      dash.time = timeSrv.timeRange();

      // remove panel queries & links
      _.each(dash.panels, panel => {
        panel.targets = [];
        panel.links = [];
        panel.datasource = null;
      });

      // remove annotation queries
      dash.annotations.list = _.chain(dash.annotations.list)
        .filter(annotation => {
          return annotation.enable;
        })
        .map((annotation: any) => {
          return {
            name: annotation.name,
            enable: annotation.enable,
            iconColor: annotation.iconColor,
            snapshotData: annotation.snapshotData,
            type: annotation.type,
            builtIn: annotation.builtIn,
            hide: annotation.hide,
          };
        })
        .value();

      // remove template queries
      _.each(dash.templating.list, variable => {
        variable.query = '';
        variable.options = variable.current;
        variable.refresh = false;
      });

      // snapshot single panel
      if ($scope.modeSharePanel) {
        const singlePanel = $scope.panel.getSaveModel();
        singlePanel.gridPos.w = 24;
        singlePanel.gridPos.x = 0;
        singlePanel.gridPos.y = 0;
        singlePanel.gridPos.h = 20;
        dash.panels = [singlePanel];
      }

      // cleanup snapshotData
      delete $scope.dashboard.snapshot;
      $scope.dashboard.forEachPanel(panel => {
        delete panel.snapshotData;
      });
      _.each($scope.dashboard.annotations.list, annotation => {
        delete annotation.snapshotData;
      });
    };

    $scope.deleteSnapshot = () => {
      backendSrv.get($scope.deleteUrl).then(() => {
        $scope.step = 3;
      });
    };
  }
}

angular.module('grafana.controllers').controller('ShareSnapshotCtrl', ShareSnapshotCtrl);
