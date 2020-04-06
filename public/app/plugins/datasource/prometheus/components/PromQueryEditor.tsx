import _ from 'lodash';
import React, { PureComponent } from 'react';

// Types
import { FormLabel, Select, Switch, QueryEditorProps, DataSourceStatus } from '@grafana/ui';
import { SelectableValue } from '@grafana/data';

import { PrometheusDatasource } from '../datasource';
import { PromQuery, PromOptions } from '../types';

import PromQueryField from './PromQueryField';
import PromLink from './PromLink';
export type Props = QueryEditorProps<PrometheusDatasource, PromQuery, PromOptions>;

const FORMAT_OPTIONS: Array<SelectableValue<string>> = [
  { label: '时间序列', value: 'time_series' },
  { label: '表', value: 'table' },
  { label: '热图', value: 'heatmap' },
];

const INTERVAL_FACTOR_OPTIONS: Array<SelectableValue<number>> = _.map([1, 2, 3, 4, 5, 10], (value: number) => ({
  value,
  label: '1/' + value,
}));

interface State {
  legendFormat: string;
  formatOption: SelectableValue<string>;
  interval: string;
  intervalFactorOption: SelectableValue<number>;
  instant: boolean;
}

export class PromQueryEditor extends PureComponent<Props, State> {
  // Query target to be modified and used for queries
  query: PromQuery;

  constructor(props: Props) {
    super(props);
    const { query } = props;
    this.query = query;
    // Query target properties that are fullu controlled inputs
    this.state = {
      // Fully controlled text inputs
      interval: query.interval,
      legendFormat: query.legendFormat,
      // Select options
      formatOption: FORMAT_OPTIONS.find(option => option.value === query.format) || FORMAT_OPTIONS[0],
      intervalFactorOption:
        INTERVAL_FACTOR_OPTIONS.find(option => option.value === query.intervalFactor) || INTERVAL_FACTOR_OPTIONS[0],
      // Switch options
      instant: Boolean(query.instant),
    };
  }

  onFieldChange = (query: PromQuery, override?: any) => {
    this.query.expr = query.expr;
  };

  onFormatChange = (option: SelectableValue<string>) => {
    this.query.format = option.value;
    this.setState({ formatOption: option }, this.onRunQuery);
  };

  onInstantChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const instant = e.target.checked;
    this.query.instant = instant;
    this.setState({ instant }, this.onRunQuery);
  };

  onIntervalChange = (e: React.SyntheticEvent<HTMLInputElement>) => {
    const interval = e.currentTarget.value;
    this.query.interval = interval;
    this.setState({ interval });
  };

  onIntervalFactorChange = (option: SelectableValue<number>) => {
    this.query.intervalFactor = option.value;
    this.setState({ intervalFactorOption: option }, this.onRunQuery);
  };

  onLegendChange = (e: React.SyntheticEvent<HTMLInputElement>) => {
    const legendFormat = e.currentTarget.value;
    this.query.legendFormat = legendFormat;
    this.setState({ legendFormat });
  };

  onRunQuery = () => {
    const { query } = this;
    this.props.onChange(query);
    this.props.onRunQuery();
  };

  render() {
    const { datasource, query, panelData, queryResponse } = this.props;
    const { formatOption, instant, interval, intervalFactorOption, legendFormat } = this.state;

    return (
      <div>
        <PromQueryField
          datasource={datasource}
          query={query}
          onRunQuery={this.onRunQuery}
          onChange={this.onFieldChange}
          history={[]}
          panelData={panelData}
          queryResponse={queryResponse}
          datasourceStatus={DataSourceStatus.Connected} // TODO: replace with real DataSourceStatus
        />

        <div className="gf-form-inline">
          <div className="gf-form">
            <FormLabel
              width={7}
              tooltip="使用名称或模式控制时间系列的名称。 例如{{hostname}}将替换为标签主机名的标签值。"
            >
              图标
            </FormLabel>
            <input
              type="text"
              className="gf-form-input"
              placeholder="图标格式"
              value={legendFormat}
              onChange={this.onLegendChange}
            />
          </div>

          <div className="gf-form">
            <FormLabel
              width={7}
              tooltip="根据时间范围和面板宽度留空以进行自动处理。请注意，将调整查询中使用的实际日期到间隔步长的倍数。"
            >
              最小步骤
            </FormLabel>
            <input
              type="text"
              className="gf-form-input width-8"
              placeholder={interval}
              onChange={this.onIntervalChange}
              value={interval}
            />
          </div>

          <div className="gf-form">
            <div className="gf-form-label">解析度</div>
            <Select
              isSearchable={false}
              options={INTERVAL_FACTOR_OPTIONS}
              onChange={this.onIntervalFactorChange}
              value={intervalFactorOption}
            />
          </div>

          <div className="gf-form">
            <div className="gf-form-label">格式化</div>
            <Select isSearchable={false} options={FORMAT_OPTIONS} onChange={this.onFormatChange} value={formatOption} />
            <Switch label="瞬间" checked={instant} onChange={this.onInstantChange} />

            <FormLabel width={10} tooltip="链接到普罗米修斯的图表">
              <PromLink
                datasource={datasource}
                query={this.query} // Use modified query
                panelData={panelData}
              />
            </FormLabel>
          </div>
        </div>
      </div>
    );
  }
}
