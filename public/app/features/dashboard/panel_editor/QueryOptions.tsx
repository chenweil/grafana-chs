// Libraries
import React, { PureComponent, ChangeEvent, FocusEvent } from 'react';

// Utils
import { rangeUtil } from '@grafana/data';

// Components
import {
  DataSourceSelectItem,
  EventsWithValidation,
  Input,
  InputStatus,
  Switch,
  ValidationEvents,
  FormLabel,
} from '@grafana/ui';
import { DataSourceOption } from './DataSourceOption';

// Types
import { PanelModel } from '../state';

const timeRangeValidationEvents: ValidationEvents = {
  [EventsWithValidation.onBlur]: [
    {
      rule: value => {
        if (!value) {
          return true;
        }
        return rangeUtil.isValidTimeSpan(value);
      },
      errorMessage: '不是有效的时间跨度',
    },
  ],
};

const emptyToNull = (value: string) => {
  return value === '' ? null : value;
};

interface Props {
  panel: PanelModel;
  datasource: DataSourceSelectItem;
}

interface State {
  relativeTime: string;
  timeShift: string;
  cacheTimeout: string;
  maxDataPoints: string;
  interval: string;
  hideTimeOverride: boolean;
}

export class QueryOptions extends PureComponent<Props, State> {
  allOptions = {
    cacheTimeout: {
      label: '缓存超时',
      placeholder: '60',
      name: 'cacheTimeout',
      tooltipInfo: (
        <>
          如果您的时间序列存储具有查询缓存，则此选项可以覆盖默认缓存超时。 指定一个以秒为单位的数值。
        </>
      ),
    },
    maxDataPoints: {
      label: '最大数据点',
      placeholder: 'auto',
      name: 'maxDataPoints',
      tooltipInfo: (
        <>
          查询应返回的最大数据点。 对于图表，这会自动设置为每个数据点一个像素。
        </>
      ),
    },
    minInterval: {
      label: '最小时间间隔',
      placeholder: '0',
      name: 'minInterval',
      panelKey: 'interval',
      tooltipInfo: (
        <>
          按时间间隔的自动组的下限。 例如，建议设置为写入频率{' '}
          <code>1m</code> 如果您的数据每分钟写一次。 通过变量访问自动间隔{' '}
          <code>$__interval</code> 对于时间范围字符串和 <code>$__interval_ms</code> 对于可以的数值变量
          用于数学表达式。
        </>
      ),
    },
  };

  constructor(props) {
    super(props);

    this.state = {
      relativeTime: props.panel.timeFrom || '',
      timeShift: props.panel.timeShift || '',
      cacheTimeout: props.panel.cacheTimeout || '',
      maxDataPoints: props.panel.maxDataPoints || '',
      interval: props.panel.interval || '',
      hideTimeOverride: props.panel.hideTimeOverride || false,
    };
  }

  onRelativeTimeChange = (event: ChangeEvent<HTMLInputElement>) => {
    this.setState({
      relativeTime: event.target.value,
    });
  };

  onTimeShiftChange = (event: ChangeEvent<HTMLInputElement>) => {
    this.setState({
      timeShift: event.target.value,
    });
  };

  onOverrideTime = (event: FocusEvent<HTMLInputElement>, status: InputStatus) => {
    const { value } = event.target;
    const { panel } = this.props;
    const emptyToNullValue = emptyToNull(value);
    if (status === InputStatus.Valid && panel.timeFrom !== emptyToNullValue) {
      panel.timeFrom = emptyToNullValue;
      panel.refresh();
    }
  };

  onTimeShift = (event: FocusEvent<HTMLInputElement>, status: InputStatus) => {
    const { value } = event.target;
    const { panel } = this.props;
    const emptyToNullValue = emptyToNull(value);
    if (status === InputStatus.Valid && panel.timeShift !== emptyToNullValue) {
      panel.timeShift = emptyToNullValue;
      panel.refresh();
    }
  };

  onToggleTimeOverride = () => {
    const { panel } = this.props;
    this.setState({ hideTimeOverride: !this.state.hideTimeOverride }, () => {
      panel.hideTimeOverride = this.state.hideTimeOverride;
      panel.refresh();
    });
  };

  onDataSourceOptionBlur = (panelKey: string) => () => {
    const { panel } = this.props;

    panel[panelKey] = this.state[panelKey];
    panel.refresh();
  };

  onDataSourceOptionChange = (panelKey: string) => (event: ChangeEvent<HTMLInputElement>) => {
    this.setState({ ...this.state, [panelKey]: event.target.value });
  };

  renderOptions = () => {
    const { datasource } = this.props;
    const { queryOptions } = datasource.meta;

    if (!queryOptions) {
      return null;
    }

    return Object.keys(queryOptions).map(key => {
      const options = this.allOptions[key];
      const panelKey = options.panelKey || key;
      return (
        <DataSourceOption
          key={key}
          {...options}
          onChange={this.onDataSourceOptionChange(panelKey)}
          onBlur={this.onDataSourceOptionBlur(panelKey)}
          value={this.state[panelKey]}
        />
      );
    });
  };

  render() {
    const { hideTimeOverride } = this.state;
    const { relativeTime, timeShift } = this.state;
    return (
      <div className="gf-form-inline">
        {this.renderOptions()}

        <div className="gf-form">
          <FormLabel>相对时间</FormLabel>
          <Input
            type="text"
            className="width-6"
            placeholder="1h"
            onChange={this.onRelativeTimeChange}
            onBlur={this.onOverrideTime}
            validationEvents={timeRangeValidationEvents}
            hideErrorMessage={true}
            value={relativeTime}
          />
        </div>

        <div className="gf-form">
          <span className="gf-form-label">时见偏移</span>
          <Input
            type="text"
            className="width-6"
            placeholder="1h"
            onChange={this.onTimeShiftChange}
            onBlur={this.onTimeShift}
            validationEvents={timeRangeValidationEvents}
            hideErrorMessage={true}
            value={timeShift}
          />
        </div>
        {(timeShift || relativeTime) && (
          <div className="gf-form-inline">
            <Switch label="隐藏时间信息" checked={hideTimeOverride} onChange={this.onToggleTimeOverride} />
          </div>
        )}
      </div>
    );
  }
}
