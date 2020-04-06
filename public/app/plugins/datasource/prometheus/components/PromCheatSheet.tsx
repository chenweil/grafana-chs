import React from 'react';

const CHEAT_SHEET_ITEMS = [
  {
    title: '请求率',
    expression: 'rate(http_request_total[5m])',
    label:
      '给定HTTP请求计数器，此查询计算过去5分钟内的每秒平均请求率。',
  },
  {
    title: '请求延迟的95%',
    expression: 'histogram_quantile(0.95, sum(rate(prometheus_http_request_duration_seconds_bucket[5m])) by (le))',
    label: '计算5分钟窗口内HTTP请求率的95%。',
  },
  {
    title: '警报',
    expression: 'sort_desc(sum(sum_over_time(ALERTS{alertstate="firing"}[24h])) by (alertname))',
    label: '统计过去24小时内发出的警报。',
  },
];

export default (props: any) => (
  <div>
    <h2>PromQL备忘单</h2>
    {CHEAT_SHEET_ITEMS.map(item => (
      <div className="cheat-sheet-item" key={item.expression}>
        <div className="cheat-sheet-item__title">{item.title}</div>
        <div
          className="cheat-sheet-item__expression"
          onClick={e => props.onClickExample({ refId: 'A', expr: item.expression })}
        >
          <code>{item.expression}</code>
        </div>
        <div className="cheat-sheet-item__label">{item.label}</div>
      </div>
    ))}
  </div>
);
