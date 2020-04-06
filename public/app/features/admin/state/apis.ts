import { getBackendSrv } from '@grafana/runtime';

export interface ServerStat {
  name: string;
  value: number;
}

export const getServerStats = async (): Promise<ServerStat[]> => {
  try {
    const res = await getBackendSrv().get('api/admin/stats');
    return [
      { name: '共计用户', value: res.users },
      { name: '共计管理员', value: res.admins },
      { name: '共计编辑数', value: res.editors },
      { name: '共计浏览数', value: res.viewers },
      { name: '激活用户数 (最近 30 天)', value: res.activeUsers },
      { name: '共计管理员 (最近 30 天)', value: res.activeAdmins },
      { name: '共计编辑者 (最近 30 天)', value: res.activeEditors },
      { name: '共计浏览者 (最近 30 天)', value: res.activeViewers },
      { name: '共计会话', value: res.activeSessions },
      { name: '共计仪表板', value: res.dashboards },
      { name: '共计组指数', value: res.orgs },
      { name: '共计播放列表', value: res.playlists },
      { name: '共计快照数', value: res.snapshots },
      { name: '共计仪表板标签数', value: res.tags },
      { name: '共计收藏仪表板', value: res.stars },
      { name: '共计提示', value: res.alerts },
    ];
  } catch (error) {
    console.error(error);
    throw error;
  }
};
