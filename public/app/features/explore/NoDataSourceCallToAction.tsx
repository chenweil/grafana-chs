import React, { useContext } from 'react';
import { css } from 'emotion';
import { ThemeContext, LinkButton, CallToActionCard } from '@grafana/ui';

export const NoDataSourceCallToAction = () => {
  const theme = useContext(ThemeContext);

  const message =
    'Explore至少需要一个数据源。 添加数据源后，可以在此处查询。';
  const footer = (
    <>
      <i className="fa fa-rocket" />
      <> ProTip: 您还可以通过配置文件定义数据源 </>
      <a
        href="http://docs.grafana.org/administration/provisioning/#datasources?utm_source=explore"
        target="_blank"
        className="text-link"
      >
        Learn more
      </a>
    </>
  );

  const ctaElement = (
    <LinkButton size="lg" href="/datasources/new" icon="gicon gicon-datasources">
      添加数据源
    </LinkButton>
  );

  const cardClassName = css`
    max-width: ${theme.breakpoints.lg};
  `;

  return (
    <CallToActionCard
      callToActionElement={ctaElement}
      className={cardClassName}
      footer={footer}
      message={message}
      theme={theme}
    />
  );
};
