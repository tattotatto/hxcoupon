import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button, Row, Col, Card, Typography, Space, Layout } from 'antd';
import {
  GiftOutlined,
  ApiOutlined,
  LinkOutlined,
  BarChartOutlined,
  RocketOutlined,
  SafetyOutlined,
  SyncOutlined,
  GlobalOutlined,
} from '@ant-design/icons';
import { useAuthStore } from '../stores/authStore';

const { Title, Text, Paragraph } = Typography;

export default function Landing() {
  const navigate = useNavigate();
  const isLoggedIn = useAuthStore((s) => s.isLoggedIn);
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 60);
    window.addEventListener('scroll', onScroll, { passive: true });
    return () => window.removeEventListener('scroll', onScroll);
  }, []);

  const navStyle: React.CSSProperties = {
    position: 'fixed',
    top: 0,
    left: 0,
    right: 0,
    zIndex: 1000,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    height: 64,
    padding: '0 48px',
    transition: 'all 0.3s ease',
    background: scrolled ? 'rgba(255,255,255,0.98)' : 'transparent',
    boxShadow: scrolled ? '0 2px 8px rgba(0,0,0,0.08)' : 'none',
  };

  const logoStyle: React.CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    gap: 10,
    cursor: 'pointer',
  };

  return (
    <div style={{ minHeight: '100vh', background: '#fff' }}>
      {/* ── Navbar ── */}
      <div style={navStyle}>
        <div style={logoStyle} onClick={() => navigate('/')}>
          <GiftOutlined style={{ fontSize: 28, color: scrolled ? '#667eea' : '#fff' }} />
          <Text strong style={{ fontSize: 20, color: scrolled ? '#333' : '#fff' }}>
            宏曦优惠券
          </Text>
        </div>
        <Space size="middle">
          {isLoggedIn ? (
            <>
              <Button type="primary" onClick={() => navigate('/admin')}>
                进入管理后台
              </Button>
              <Button
                ghost={!scrolled}
                type={scrolled ? 'default' : 'default'}
                style={scrolled ? {} : { color: '#fff', borderColor: '#fff' }}
                onClick={() => window.open('/docs', '_blank')}
              >
                API 文档
              </Button>
            </>
          ) : (
            <>
              <Button
                type="text"
                style={{ color: scrolled ? '#333' : '#fff', fontSize: 15 }}
                onClick={() => navigate('/login')}
              >
                登录
              </Button>
              <Button
                ghost={!scrolled}
                style={scrolled ? {} : { color: '#fff', borderColor: '#fff' }}
                onClick={() => navigate('/register')}
              >
                注册
              </Button>
              <Button
                type={scrolled ? 'primary' : 'default'}
                ghost={!scrolled}
                style={scrolled ? {} : { color: '#fff', borderColor: '#fff' }}
                onClick={() => window.open('/docs', '_blank')}
              >
                API 文档
              </Button>
            </>
          )}
        </Space>
      </div>

      {/* ── Hero ── */}
      <div
        style={{
          minHeight: '100vh',
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'center',
          alignItems: 'center',
          background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
          padding: '120px 48px 80px',
          textAlign: 'center',
        }}
      >
        <Title
          style={{
            fontSize: 52,
            fontWeight: 800,
            color: '#fff',
            marginBottom: 16,
            letterSpacing: 2,
          }}
        >
          宏曦优惠券管理平台
        </Title>
        <Title
          level={2}
          style={{
            fontWeight: 400,
            color: 'rgba(255,255,255,0.9)',
            marginBottom: 24,
            maxWidth: 720,
          }}
        >
          通用、开放、高效的优惠券解决方案
          <br />
          助力您的平台提升用户粘性与复购率
        </Title>
        <Paragraph
          style={{
            fontSize: 17,
            color: 'rgba(255,255,255,0.75)',
            maxWidth: 640,
            marginBottom: 40,
            lineHeight: 1.8,
          }}
        >
          无论是电商、SaaS 还是线下门店，只需接入一套 API，即可拥有完整的优惠券发券、核销与数据分析能力。
          平台无关、语言无关，基于 HMAC 签名的安全开放接口，让集成变得前所未有地简单。
        </Paragraph>
        <Space size="large">
          <Button
            type="primary"
            size="large"
            icon={<RocketOutlined />}
            style={{
              height: 48,
              paddingLeft: 32,
              paddingRight: 32,
              fontSize: 16,
              borderRadius: 8,
              background: '#fff',
              color: '#667eea',
              border: 'none',
              fontWeight: 600,
            }}
            onClick={() => navigate(isLoggedIn ? '/admin' : '/register')}
          >
            {isLoggedIn ? '进入管理后台' : '免费注册'}
          </Button>
          <Button
            size="large"
            icon={<ApiOutlined />}
            ghost
            style={{
              height: 48,
              paddingLeft: 32,
              paddingRight: 32,
              fontSize: 16,
              borderRadius: 8,
              color: '#fff',
              borderColor: '#fff',
            }}
            onClick={() => window.open('/docs', '_blank')}
          >
            查看 API 文档
          </Button>
        </Space>
      </div>

      {/* ── Features ── */}
      <div style={{ padding: '80px 48px', maxWidth: 1200, margin: '0 auto' }}>
        <div style={{ textAlign: 'center', marginBottom: 56 }}>
          <Title level={2} style={{ marginBottom: 12, fontWeight: 700 }}>
            为什么选择宏曦？
          </Title>
          <Text type="secondary" style={{ fontSize: 16 }}>
            一套平台，满足所有优惠券业务需求
          </Text>
        </div>
        <Row gutter={[32, 32]}>
          <Col xs={24} sm={12} lg={6}>
            <Card
              hoverable
              style={{ borderRadius: 12, height: '100%', textAlign: 'center', paddingTop: 16 }}
              bodyStyle={{ padding: '24px 20px 32px' }}
            >
              <GlobalOutlined style={{ fontSize: 40, color: '#667eea', marginBottom: 16 }} />
              <Title level={4}>通用发券平台</Title>
              <Text type="secondary">
                适配各类业务场景，无论是电商促销、会员权益、积分兑换还是线下活动，
                一套 API 覆盖所有发券需求。
              </Text>
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card
              hoverable
              style={{ borderRadius: 12, height: '100%', textAlign: 'center', paddingTop: 16 }}
              bodyStyle={{ padding: '24px 20px 32px' }}
            >
              <LinkOutlined style={{ fontSize: 40, color: '#52c41a', marginBottom: 16 }} />
              <Title level={4}>平台无关性</Title>
              <Text type="secondary">
                基于 HMAC 签名的开放 API，不限制开发语言和框架。
                Java、Python、Go、PHP、C# 均可轻松对接，真正的平台无关。
              </Text>
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card
              hoverable
              style={{ borderRadius: 12, height: '100%', textAlign: 'center', paddingTop: 16 }}
              bodyStyle={{ padding: '24px 20px 32px' }}
            >
              <SyncOutlined style={{ fontSize: 40, color: '#fa8c16', marginBottom: 16 }} />
              <Title level={4}>完整生命周期</Title>
              <Text type="secondary">
                从券模板设计、批量发放、扫码核销到退券退款，
                覆盖优惠券全流程管理，无需额外开发。
              </Text>
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card
              hoverable
              style={{ borderRadius: 12, height: '100%', textAlign: 'center', paddingTop: 16 }}
              bodyStyle={{ padding: '24px 20px 32px' }}
            >
              <BarChartOutlined style={{ fontSize: 40, color: '#ff4d4f', marginBottom: 16 }} />
              <Title level={4}>数据驱动增长</Title>
              <Text type="secondary">
                实时仪表盘查看发放与核销趋势，用数据衡量活动效果，
                持续优化优惠券策略，提升用户粘性与复购率。
              </Text>
            </Card>
          </Col>
        </Row>
      </div>

      {/* ── How It Works ── */}
      <div style={{ padding: '80px 48px', background: '#f7f8fc' }}>
        <div style={{ maxWidth: 1000, margin: '0 auto' }}>
          <div style={{ textAlign: 'center', marginBottom: 48 }}>
            <Title level={2} style={{ marginBottom: 12, fontWeight: 700 }}>
              三步接入，即刻发券
            </Title>
            <Text type="secondary" style={{ fontSize: 16 }}>
              从注册到发券，最快只需 5 分钟
            </Text>
          </div>
          <Row gutter={[48, 32]}>
            <Col xs={24} md={8}>
              <div style={{ textAlign: 'center' }}>
                <div
                  style={{
                    width: 72,
                    height: 72,
                    borderRadius: 36,
                    background: 'linear-gradient(135deg, #667eea, #764ba2)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    margin: '0 auto 20px',
                  }}
                >
                  <span style={{ color: '#fff', fontSize: 28, fontWeight: 700 }}>1</span>
                </div>
                <Title level={4}>注册账号</Title>
                <Text type="secondary">
                  提交商家信息，获取 AppKey 和 AppSecret，即可开始对接 API
                </Text>
              </div>
            </Col>
            <Col xs={24} md={8}>
              <div style={{ textAlign: 'center' }}>
                <div
                  style={{
                    width: 72,
                    height: 72,
                    borderRadius: 36,
                    background: 'linear-gradient(135deg, #667eea, #764ba2)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    margin: '0 auto 20px',
                  }}
                >
                  <span style={{ color: '#fff', fontSize: 28, fontWeight: 700 }}>2</span>
                </div>
                <Title level={4}>创建模板 & 发券</Title>
                <Text type="secondary">
                  在管理后台创建优惠券模板，通过 API 或后台一键发放
                </Text>
              </div>
            </Col>
            <Col xs={24} md={8}>
              <div style={{ textAlign: 'center' }}>
                <div
                  style={{
                    width: 72,
                    height: 72,
                    borderRadius: 36,
                    background: 'linear-gradient(135deg, #667eea, #764ba2)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    margin: '0 auto 20px',
                  }}
                >
                  <span style={{ color: '#fff', fontSize: 28, fontWeight: 700 }}>3</span>
                </div>
                <Title level={4}>核销 & 分析</Title>
                <Text type="secondary">
                  用户到店核销优惠券，后台实时统计分析活动效果
                </Text>
              </div>
            </Col>
          </Row>
        </div>
      </div>

      {/* ── Footer ── */}
      <Layout.Footer style={{ textAlign: 'center', background: '#1a1a2e', color: 'rgba(255,255,255,0.6)', padding: '32px 48px' }}>
        <div style={{ marginBottom: 8 }}>
          <Space>
            <GiftOutlined style={{ color: 'rgba(255,255,255,0.4)' }} />
            <Text style={{ color: 'rgba(255,255,255,0.6)' }}>宏曦优惠券管理平台</Text>
          </Space>
        </div>
        <div>
          <Text style={{ color: 'rgba(255,255,255,0.4)', fontSize: 13 }}>
            © {new Date().getFullYear()} 宏曦优惠券 | 通用开放优惠券解决方案
          </Text>
        </div>
      </Layout.Footer>
    </div>
  );
}
