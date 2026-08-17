import { Card, Button } from 'antd';
import React from 'react';
import { useNavigate } from 'react-router-dom';
import type { QuickLinksPanelConfig } from './dashboard-config';
import * as icons from '@ant-design/icons';
import { colors, fonts, radius } from '../../theme/tokens';

interface QuickLinksPanelProps {
  config: QuickLinksPanelConfig;
}

export default function QuickLinksPanel({ config }: QuickLinksPanelProps) {
  const navigate = useNavigate();

  return (
    <Card
      title={config.title}
      style={{
        background: colors.bg.panel,
        border: 'none',
        borderRadius: radius.panel,
      }}
      headStyle={{
        color: colors.text.primary,
        borderBottom: `1px solid ${colors.border}`,
        fontSize: fonts.size.h2,
        fontWeight: fonts.weight.h2,
      }}
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        {config.links.map((link) => {
          const IconComponent = (icons as unknown as Record<string, React.ComponentType>)[link.icon];
          return (
            <Button
              key={link.path}
              type="default"
              icon={IconComponent ? React.createElement(IconComponent) : undefined}
              onClick={() => navigate(link.path)}
              style={{
                background: 'transparent',
                color: colors.text.secondary,
                border: `1px solid ${colors.borderLight}`,
                textAlign: 'left',
                height: 40,
              }}
            >
              {link.label}
            </Button>
          );
        })}
      </div>
    </Card>
  );
}
