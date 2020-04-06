import coreModule from '../../core_module';
import appEvents from 'app/core/app_events';

export class HelpCtrl {
  tabIndex: any;
  shortcuts: any;

  /** @ngInject */
  constructor() {
    this.tabIndex = 0;
    this.shortcuts = {
      '全局': [
        { keys: ['g', 'h'], description: '返回主面板' },
        { keys: ['g', 'p'], description: '返回参数' },
        { keys: ['s', 'o'], description: '打开搜索' },
        { keys: ['esc'], description: '推出编辑/查看设置' },
      ],
      '仪表板': [
        { keys: ['mod+s'], description: '保存仪表板' },
        { keys: ['d', 'r'], description: '刷新所有面板' },
        { keys: ['d', 's'], description: '仪表板设置' },
        { keys: ['d', 'v'], description: '切换活动/查看模式' },
        { keys: ['d', 'k'], description: '切换信息展亭模式（隐藏顶部导航）' },
        { keys: ['d', 'E'], description: '展开所有行' },
        { keys: ['d', 'C'], description: '折叠所有行' },
        { keys: ['d', 'a'], description: '切换自动配合面板（实验功能）' },
        { keys: ['mod+o'], description: '切换共享图形十字准线' },
        { keys: ['d', 'l'], description: '切换所有面板图例' },
      ],
      '展示面板': [
        { keys: ['e'], description: '切换面板编辑视图' },
        { keys: ['v'], description: '切换面板全屏视图' },
        { keys: ['p', 's'], description: '打开面板共享模式' },
        { keys: ['p', 'd'], description: '复制面板' },
        { keys: ['p', 'r'], description: '删除面板' },
        { keys: ['p', 'l'], description: '切换面板图例' },
      ],
      '时间范围': [
        { keys: ['t', 'z'], description: '缩小时间范围' },
        {
          keys: ['t', '<i class="fa fa-long-arrow-left"></i>'],
          description: '移回时间范围',
        },
        {
          keys: ['t', '<i class="fa fa-long-arrow-right"></i>'],
          description: '向前移动时间范围',
        },
      ],
    };
  }

  dismiss() {
    appEvents.emit('hide-modal');
  }
}

export function helpModal() {
  return {
    restrict: 'E',
    templateUrl: 'public/app/core/components/help/help.html',
    controller: HelpCtrl,
    bindToController: true,
    transclude: true,
    controllerAs: 'ctrl',
    scope: {},
  };
}

coreModule.directive('helpModal', helpModal);
