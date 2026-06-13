import { useState, useEffect, useCallback } from 'react';
import { Row, Col, Card, Tag, Typography, Spin, Empty, Pagination, Space, Modal, Tabs, Button, message } from 'antd';
import { CopyOutlined, CodeOutlined } from '@ant-design/icons';
import { templateApi } from '../../api/template';
import dayjs from 'dayjs';

const { Title, Text, Paragraph } = Typography;

const typeMap: Record<string, { label: string; color: string }> = {
  full_reduction: { label: '满减', color: 'blue' },
  discount: { label: '折扣', color: 'purple' },
  fixed_amount: { label: '固定金额', color: 'green' },
};

const BASE_DOMAIN = 'https://coupon.mx.yn.cn';

function buildExamples(templateId: number) {
  const body = JSON.stringify({ template_id: templateId, user_phone: '13800138000' }, null, 2);
  const path = '/api/v1/coupons/issue';
  const fullUrl = `${BASE_DOMAIN}${path}`;

  const curl = `curl -X POST "${fullUrl}" \\
  -H "X-App-Key: YOUR_APP_KEY" \\
  -H "X-Timestamp: $(date +%s)" \\
  -H "X-Nonce: $(uuidgen)" \\
  -H "X-Signature: <GENERATED_SIGNATURE>" \\
  -H "Content-Type: application/json" \\
  -d '${JSON.stringify({ template_id: templateId, user_phone: '13800138000' })}'`;

  const js = `// generate HMAC headers with signRequest()
const headers = signRequest('POST', '${path}', ${JSON.stringify({ template_id: templateId, user_phone: '13800138000' })}, appKey, appSecret);

fetch('${fullUrl}', {
  method: 'POST',
  headers,
  body: JSON.stringify({ template_id: ${templateId}, user_phone: '13800138000' }),
})
  .then(r => r.json())
  .then(data => console.log(data));`;

  const py = `import requests

headers = sign_request('POST', '${path}', {"template_id": ${templateId}, "user_phone": "13800138000"}, app_key, app_secret)
resp = requests.post('${fullUrl}', json={"template_id": ${templateId}, "user_phone": "13800138000"}, headers=headers)
print(resp.json())`;

  const go = `import (
    "bytes"
    "encoding/json"
    "net/http"
)

body := map[string]interface{}{
    "template_id": ${templateId},
    "user_phone":   "13800138000",
}
bodyJSON, _ := json.Marshal(body)

req, _ := http.NewRequest("POST", "${fullUrl}", bytes.NewBuffer(bodyJSON))
// headers := SignRequest("POST", "${path}", string(bodyJSON), appKey, appSecret)
// for k, v := range headers { req.Header.Set(k, v) }
req.Header.Set("Content-Type", "application/json")

client := &http.Client{}
resp, _ := client.Do(req)
defer resp.Body.Close()
// read resp.Body...`;

  const php = `<?php
$body = json_encode(["template_id" => ${templateId}, "user_phone" => "13800138000"]);

$ch = curl_init();
curl_setopt($ch, CURLOPT_URL, '${fullUrl}');
curl_setopt($ch, CURLOPT_CUSTOMREQUEST, 'POST');
curl_setopt($ch, CURLOPT_HTTPHEADER, [
    // Use sign_request() to generate headers
    // 'X-App-Key: YOUR_APP_KEY',
    // 'X-Timestamp: ' . $timestamp,
    // 'X-Nonce: ' . $nonce,
    // 'X-Signature: ' . $signature,
    'Content-Type: application/json',
]);
curl_setopt($ch, CURLOPT_POSTFIELDS, $body);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
$response = curl_exec($ch);
curl_close($ch);

echo $response;`;

  const cs = `using System.Net.Http;
using System.Text;
using Newtonsoft.Json;

var client = new HttpClient();
var body = new { template_id = ${templateId}, user_phone = "13800138000" };
var content = new StringContent(
    JsonConvert.SerializeObject(body),
    Encoding.UTF8, "application/json");

// var headers = SignRequest("POST", "${path}", content, appKey, appSecret);
// foreach (var h in headers) client.DefaultRequestHeaders.Add(h.Key, h.Value);

var response = await client.PostAsync("${fullUrl}", content);
var result = await response.Content.ReadAsStringAsync();
Console.WriteLine(result);`;

  const java = `import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import java.net.URI;
import java.net.http.*;
import java.time.Instant;
import java.util.Base64;
import java.util.UUID;

var client = HttpClient.newHttpClient();
var body = """
{
  "template_id": ${templateId},
  "user_phone": "13800138000"
}
""";

var timestamp = String.valueOf(Instant.now().getEpochSecond());
var nonce = UUID.randomUUID().toString();
var signingStr = "POST" + "\\n" + "${path}" + "\\n"
    + timestamp + "\\n" + nonce + "\\n" + body;

var mac = Mac.getInstance("HmacSHA256");
mac.init(new SecretKeySpec(appSecret.getBytes(), "HmacSHA256"));
var signature = Base64.getEncoder()
    .encodeToString(mac.doFinal(signingStr.getBytes()));

var request = HttpRequest.newBuilder()
    .uri(URI.create("${fullUrl}"))
    .header("X-App-Key", appKey)
    .header("X-Timestamp", timestamp)
    .header("X-Nonce", nonce)
    .header("X-Signature", signature)
    .header("Content-Type", "application/json")
    .POST(HttpRequest.BodyPublishers.ofString(body))
    .build();

var response = client.send(request, HttpResponse.BodyHandlers.ofString());
System.out.println(response.body());`;

  return { curl, js, py, go, php, cs, java, body, path };
}

export default function TemplateBrowse() {
  const [items, setItems] = useState<any[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [filters, setFilters] = useState({ page: 1, page_size: 12 });
  const [selected, setSelected] = useState<any>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await templateApi.browse(filters);
      setItems(res.data.data.items || []);
      setTotal(res.data.data.total || 0);
    } catch { /* handled */ }
    finally { setLoading(false); }
  }, [filters]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const examples = selected ? buildExamples(selected.id) : null;

  return (
    <Spin spinning={loading}>
      <div>
        <Title level={4} style={{ marginBottom: 24 }}>模板浏览 — 选择一个模板查看发券接口</Title>
        {items.length === 0 && !loading ? (
          <Empty description="暂无可用模板" />
        ) : (
          <>
            <Row gutter={[16, 16]}>
              {items.map((item) => {
                const typeInfo = typeMap[item.type] || { label: item.type, color: 'default' };
                return (
                  <Col xs={24} sm={12} lg={8} key={item.id}>
                    <Card
                      hoverable
                      style={{ borderRadius: 12 }}
                      title={
                        <Space>
                          <Tag color={typeInfo.color}>{typeInfo.label}</Tag>
                          <Text strong>{item.name}</Text>
                        </Space>
                      }
                      onClick={() => setSelected(item)}
                    >
                      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 12 }}>
                        <div>
                          <Text type="secondary" style={{ fontSize: 12 }}>优惠</Text>
                          <div style={{ fontSize: 24, fontWeight: 600, color: '#f5222d' }}>
                            {item.type === 'discount' ? `${(item.discount_value / 10).toFixed(1)}折` : `¥${item.discount_value}`}
                          </div>
                        </div>
                        <div style={{ textAlign: 'right' }}>
                          <Text type="secondary" style={{ fontSize: 12 }}>门槛</Text>
                          <div style={{ fontSize: 24, fontWeight: 600 }}>¥{item.threshold_amount}</div>
                        </div>
                      </div>
                      <div style={{ color: '#8c8c8c', fontSize: 13 }}>
                        {item.store_name && <div>门店: {item.store_name}</div>}
                        <div>库存: {item.total_quantity}</div>
                        <div>有效期: {item.validity_type === 'days_after_receive'
                          ? `领取后${item.validity_days}天`
                          : `${item.valid_start ? dayjs(item.valid_start).format('MM/DD') : ''} - ${item.valid_end ? dayjs(item.valid_end).format('MM/DD') : ''}`}</div>
                        <div style={{ marginTop: 4, color: '#1677ff', fontSize: 12 }}>点击查看发券接口 →</div>
                      </div>
                    </Card>
                  </Col>
                );
              })}
            </Row>
            {total > filters.page_size && (
              <div style={{ textAlign: 'center', marginTop: 24 }}>
                <Pagination
                  current={filters.page}
                  pageSize={filters.page_size}
                  total={total}
                  onChange={(p) => setFilters((f) => ({ ...f, page: p }))}
                />
              </div>
            )}
          </>
        )}

        {/* Issue API Guide Modal */}
        <Modal
          open={!!selected}
          onCancel={() => setSelected(null)}
          title={selected ? `发券接口 — ${selected.name}` : ''}
          width={700}
          footer={null}
        >
          {selected && examples && (
            <div>
              {/* Template Summary */}
              <Card size="small" style={{ marginBottom: 16, background: '#f0f5ff' }}>
                <Space wrap>
                  {selected.store_name && <Tag color="green">{selected.store_name}</Tag>}
                  <Tag color={typeMap[selected.type]?.color}>{typeMap[selected.type]?.label}</Tag>
                  <Text strong>优惠:</Text>
                  <Text style={{ color: '#f5222d', fontSize: 18, fontWeight: 600 }}>
                    {selected.type === 'discount' ? `${(selected.discount_value / 10).toFixed(1)}折` : `¥${selected.discount_value}`}
                  </Text>
                  <Text strong>门槛: ¥{selected.threshold_amount}</Text>
                  <Text strong>模板ID:</Text>
                  <Tag>{selected.id}</Tag>
                </Space>
                {selected.mp_appid && (
                  <div style={{ marginTop: 8 }}>
                    <Text type="secondary">跳转小程序: </Text>
                    <Text code>{selected.mp_appid}</Text>
                    <Text type="secondary"> 路径: </Text>
                    <Text code>{selected.mp_page_path || '/'}</Text>
                  </div>
                )}
              </Card>

              {/* API Info */}
              <div style={{ marginBottom: 16 }}>
                <Space>
                  <Tag color="#1677ff" style={{ fontSize: 14, padding: '4px 12px' }}>POST</Tag>
                  <Text code style={{ fontSize: 14 }}>{BASE_DOMAIN}/api/v1/coupons/issue</Text>
                  <Button size="small" icon={<CopyOutlined />}
                    onClick={() => { navigator.clipboard.writeText(`${BASE_DOMAIN}/api/v1/coupons/issue`); message.success('已复制'); }} />
                </Space>
              </div>

              <Paragraph type="secondary" style={{ marginBottom: 16 }}>
                使用你门店的 <Text code>AppKey</Text> 和 <Text code>AppSecret</Text> 生成 HMAC 签名，
                请求体中 <Text code>template_id</Text> 已预填为 <Tag>{selected.id}</Tag>，
                <Text code>user_phone</Text> 替换为领券用户的手机号。
              </Paragraph>

              {/* HMAC Auth Headers */}
              <Card size="small" title="请求头 (HMAC 签名)" style={{ marginBottom: 16 }}>
                <pre style={{ background: '#1e1e1e', color: '#d4d4d4', padding: 12, borderRadius: 6, fontSize: 12, margin: 0 }}>
{`X-App-Key: YOUR_APP_KEY
X-Timestamp: 当前Unix时间戳
X-Nonce: 随机UUID
X-Signature: Base64(HMAC-SHA256(AppSecret, 签名串))
Content-Type: application/json

签名串 = HTTP方法 + "\\n" + URL路径 + "\\n" + Timestamp + "\\n" + Nonce + "\\n" + 请求体JSON`}
                </pre>
              </Card>

              {/* Request Body */}
              <Card size="small" title="请求体" extra={
                <Button size="small" icon={<CopyOutlined />}
                  onClick={() => { navigator.clipboard.writeText(examples.body); message.success('已复制'); }}>复制</Button>
              } style={{ marginBottom: 16 }}>
                <pre style={{ background: '#f5f5f5', padding: 12, borderRadius: 6, fontSize: 13, margin: 0 }}>
                  {examples.body}
                </pre>
              </Card>

              {/* Expected Response */}
              <Card size="small" title="预期返回" style={{ marginBottom: 16 }}>
                <pre style={{ background: '#f6ffed', padding: 12, borderRadius: 6, border: '1px solid #b7eb8f', fontSize: 12, margin: 0 }}>
{`{
  "code": 0,
  "message": "success",
  "data": {
    "coupon_id": 1,
    "coupon_code": "XXX...",
    "template_name": "${selected.name}",
    "type": "${selected.type}",
    "discount_value": ${selected.discount_value},
    "threshold_amount": ${selected.threshold_amount},
    "valid_start": "...",
    "valid_end": "...",
    "status": "unused"
  }
}`}
                </pre>
              </Card>

              {/* Code Examples */}
              <Card size="small" title={<Space><CodeOutlined />代码示例</Space>}>
                <Tabs
                  size="small"
                  items={[
                    {
                      key: 'curl',
                      label: 'cURL',
                      children: (
                        <div>
                          <div style={{ textAlign: 'right', marginBottom: 4 }}>
                            <Button size="small" icon={<CopyOutlined />}
                              onClick={() => { navigator.clipboard.writeText(examples.curl); message.success('已复制'); }}>复制</Button>
                          </div>
                          <pre style={{ background: '#1e1e1e', color: '#d4d4d4', padding: 12, borderRadius: 6, fontSize: 11, lineHeight: 1.6, margin: 0, overflow: 'auto' }}>
                            {examples.curl}
                          </pre>
                        </div>
                      ),
                    },
                    {
                      key: 'js',
                      label: 'JavaScript',
                      children: (
                        <div>
                          <div style={{ textAlign: 'right', marginBottom: 4 }}>
                            <Button size="small" icon={<CopyOutlined />}
                              onClick={() => { navigator.clipboard.writeText(examples.js); message.success('已复制'); }}>复制</Button>
                          </div>
                          <pre style={{ background: '#1e1e1e', color: '#d4d4d4', padding: 12, borderRadius: 6, fontSize: 11, lineHeight: 1.6, margin: 0, overflow: 'auto' }}>
                            {examples.js}
                          </pre>
                        </div>
                      ),
                    },
                    {
                      key: 'py',
                      label: 'Python',
                      children: (
                        <div>
                          <div style={{ textAlign: 'right', marginBottom: 4 }}>
                            <Button size="small" icon={<CopyOutlined />}
                              onClick={() => { navigator.clipboard.writeText(examples.py); message.success('已复制'); }}>复制</Button>
                          </div>
                          <pre style={{ background: '#1e1e1e', color: '#d4d4d4', padding: 12, borderRadius: 6, fontSize: 11, lineHeight: 1.6, margin: 0, overflow: 'auto' }}>
                            {examples.py}
                          </pre>
                        </div>
                      ),
                    },
                    {
                      key: 'go',
                      label: 'Go',
                      children: (
                        <div>
                          <div style={{ textAlign: 'right', marginBottom: 4 }}>
                            <Button size="small" icon={<CopyOutlined />}
                              onClick={() => { navigator.clipboard.writeText(examples.go); message.success('已复制'); }}>复制</Button>
                          </div>
                          <pre style={{ background: '#1e1e1e', color: '#d4d4d4', padding: 12, borderRadius: 6, fontSize: 11, lineHeight: 1.6, margin: 0, overflow: 'auto' }}>
                            {examples.go}
                          </pre>
                        </div>
                      ),
                    },
                    {
                      key: 'php',
                      label: 'PHP',
                      children: (
                        <div>
                          <div style={{ textAlign: 'right', marginBottom: 4 }}>
                            <Button size="small" icon={<CopyOutlined />}
                              onClick={() => { navigator.clipboard.writeText(examples.php); message.success('已复制'); }}>复制</Button>
                          </div>
                          <pre style={{ background: '#1e1e1e', color: '#d4d4d4', padding: 12, borderRadius: 6, fontSize: 11, lineHeight: 1.6, margin: 0, overflow: 'auto' }}>
                            {examples.php}
                          </pre>
                        </div>
                      ),
                    },
                    {
                      key: 'cs',
                      label: 'C#',
                      children: (
                        <div>
                          <div style={{ textAlign: 'right', marginBottom: 4 }}>
                            <Button size="small" icon={<CopyOutlined />}
                              onClick={() => { navigator.clipboard.writeText(examples.cs); message.success('已复制'); }}>复制</Button>
                          </div>
                          <pre style={{ background: '#1e1e1e', color: '#d4d4d4', padding: 12, borderRadius: 6, fontSize: 11, lineHeight: 1.6, margin: 0, overflow: 'auto' }}>
                            {examples.cs}
                          </pre>
                        </div>
                      ),
                    },
                    {
                      key: 'java',
                      label: 'Java',
                      children: (
                        <div>
                          <div style={{ textAlign: 'right', marginBottom: 4 }}>
                            <Button size="small" icon={<CopyOutlined />}
                              onClick={() => { navigator.clipboard.writeText(examples.java); message.success('已复制'); }}>复制</Button>
                          </div>
                          <pre style={{ background: '#1e1e1e', color: '#d4d4d4', padding: 12, borderRadius: 6, fontSize: 11, lineHeight: 1.6, margin: 0, overflow: 'auto' }}>
                            {examples.java}
                          </pre>
                        </div>
                      ),
                    },
                  ]}
                />
              </Card>

              {/* Steps */}
              <Card size="small" title="集成步骤" style={{ marginTop: 16, background: '#fffbe6' }}>
                <ol style={{ margin: 0, paddingLeft: 20, fontSize: 13, lineHeight: 2 }}>
                  <li>在「门店管理」→「密钥」获取 <Text code>AppKey</Text> 和 <Text code>AppSecret</Text></li>
                  <li>参考 API 文档实现 HMAC 签名算法（cURL / JS / Python / Go / PHP / C# / Java）</li>
                  <li>问卷/活动页面用户提交手机号后，调用本接口发券</li>
                  {selected.mp_appid && (
                    <li>在用户优惠券列表中使用 <Text code>mp_appid</Text>={<Text code>{selected.mp_appid}</Text>} 和 <Text code>mp_page_path</Text>={<Text code>{selected.mp_page_path || '/'}</Text>} 实现「去用券」按钮跳转</li>
                  )}
                </ol>
              </Card>
            </div>
          )}
        </Modal>
      </div>
    </Spin>
  );
}
