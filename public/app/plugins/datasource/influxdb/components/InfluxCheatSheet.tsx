import React from 'react';

const CHEAT_SHEET_ITEMS = [
  {
    title: '入门',
    label:
      '通过从上方的下拉菜单中选择一个测量和字段开始。 然后，您可以使用标签选择器进一步缩小搜索范围。',
  },
];

export default (props: any) => (
  <div>
    <h2>InfluxDB Cheat Sheet</h2>
    {CHEAT_SHEET_ITEMS.map(item => (
      <div className="cheat-sheet-item" key={item.title}>
        <div className="cheat-sheet-item__title">{item.title}</div>
        <div className="cheat-sheet-item__label">{item.label}</div>
      </div>
    ))}
  </div>
);
