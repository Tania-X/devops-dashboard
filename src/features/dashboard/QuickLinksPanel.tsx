import { Card, Button } from 'antd';
import React from 'react';
import { useNavigate } from 'react-router-dom';
import type { QuickLinksPanelConfig } from './dashboard-config';
import * as icons from '@ant-design/icons';

interface QuickLinksPanelProps {
  config: QuickLinksPanelConfig;
}

export default function QuickLinksPanel({ config }: QuickLinksPanelProps) {
  const navigate = useNavigate();

  return (
    <Card
      title={config.title}
      style={{
        background: '#1f1f1f',
        border: 'none',
        borderRadius: 4,
      }}
      headStyle={{
        color: '#ffffff',
        borderBottom: '1px solid #333333',
        fontSize: 16,
        fontWeight: 500,
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
                color: '#aaaaaa',
                border: '1px solid #555',
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
