import { useEffect, useState } from "react";
import { Table, Card, Spin, Alert, InputNumber, Select, Button, Tabs } from "antd";
import { Typography } from "antd";
import OILineChart from "./charts";
const { Option } = Select;
const { Title } = Typography;
const { TabPane } = Tabs;

const columns = [
 {
    title: "Timestamp",
    dataIndex: "timestamp",
    key: "timestamp",
    width: 130,
    align: "center",
    sorter: (a, b) => {
      const dateA = new Date(a.timestamp);
      const dateB = new Date(b.timestamp);
      return dateA.getTime() - dateB.getTime();
    },
  },
  {
    title: "Expiry Date",
    dataIndex: "expiryDate",
    key: "expiryDate",
    sorter: (a, b) => a.expiryDate.localeCompare(b.expiryDate),
    width: 100,
    align: "center",
  },
  {
    title: "COI",
    dataIndex: "ceOpenInterest",
    key: "ceOpenInterest",
    render: (v) => (v !== undefined ? v.toLocaleString() : "-"),
    sorter: (a, b) => (a.ceOpenInterest || 0) - (b.ceOpenInterest || 0),
    align: "right",
    width: 80,
  },
  {
    title: "CCOI",
    dataIndex: "ceChangeInOI",
    key: "ceChangeInOI",
    render: (v) => (v !== undefined ? v.toLocaleString() : "-"),
    sorter: (a, b) => (a.ceChangeInOI || 0) - (b.ceChangeInOI || 0),
    width: 80,
    align: "right",
  },
  {
    title: "CCOI%",
    dataIndex: "cePChangeInOI",
    key: "cePChangeInOI",
    render: (v) => (v !== undefined ? v.toFixed(2) + "%" : "-"),
    sorter: (a, b) => (a.cePChangeInOI || 0) - (b.cePChangeInOI || 0),
    width: 80,
    align: "right",
  },
  {
    title: "CVol",
    dataIndex: "ceVolume",
    key: "ceVolume",
    render: (v) => (v !== undefined ? v.toLocaleString() : "-"),
    sorter: (a, b) => (a.ceVolume || 0) - (b.ceVolume || 0),
    width: 80,
    align: "right",
  },
  {
    title: "CIV",
    dataIndex: "ceIV",
    key: "ceIV",
    render: (v) => (v !== undefined ? v.toFixed(2) : "-"),
    sorter: (a, b) => (a.ceIV || 0) - (b.ceIV || 0),
    width: 70,
    align: "right",
  },
  {
    title: "CE LTP",
    dataIndex: "ceLTP",
    key: "ceLTP",
    render: (v) => (v !== undefined ? v.toFixed(2) : "-"),
    sorter: (a, b) => (a.ceLTP || 0) - (b.ceLTP || 0),
    width: 80,
    align: "right",
  },
  {
    title: "Spot",
    dataIndex: "underlyingValue",
    key: "underlyingValue",
    render: (v) => v ? v.toFixed(2) : "-",
    sorter: (a, b) => (a.underlyingValue || 0) - (b.underlyingValue || 0),
    width: 80,
    align: "center",
  },
  {
    title: <span>Strike Price</span>,
    dataIndex: "strikePrice",
    key: "strikePrice",
    render: (v) => <div className="bg-amber-200 p-0"><strong className="text-amber-600 p-0.5">{v?.toLocaleString()}</strong></div>,
    sorter: (a, b) => (a.strikePrice || 0) - (b.strikePrice || 0),
    width: 100,
    align: "center",
  },
  {
    title: "PE LTP",
    dataIndex: "peLTP",
    key: "peLTP",
    render: (v) => (v !== undefined ? v.toFixed(2) : "-"),
    sorter: (a, b) => (a.peLTP || 0) - (b.peLTP || 0),
    width: 80,
    align: "right",
  },
  {
    title: "PE IV",
    dataIndex: "peIV",
    key: "peIV",
    render: (v) => (v !== undefined ? v.toFixed(2) : "-"),
    sorter: (a, b) => (a.peIV || 0) - (b.peIV || 0),
    width: 70,
    align: "right",
  },
  {
    title: "PE Vol",
    dataIndex: "peVolume",
    key: "peVolume",
    render: (v) => (v !== undefined ? v.toLocaleString() : "-"),
    sorter: (a, b) => (a.peVolume || 0) - (b.peVolume || 0),
    width: 80,
    align: "right",
  },
  {
    title: "PCOI%",
    dataIndex: "pePChangeInOI",
    key: "pePChangeInOI",
    render: (v) => (v !== undefined ? v.toFixed(2) + "%" : "-"),
    sorter: (a, b) => (a.pePChangeInOI || 0) - (b.pePChangeInOI || 0),
    width: 80,
    align: "right",
  },
  {
    title: "PCOI",
    dataIndex: "peChangeInOI",
    key: "peChangeInOI",
    render: (v) => (v !== undefined ? v.toLocaleString() : "-"),
    sorter: (a, b) => (a.peChangeInOI || 0) - (b.peChangeInOI || 0),
    width: 80,
    align: "right",
  },
  {
    title: "POI",
    dataIndex: "peOpenInterest",
    key: "peOpenInterest",
    render: (v) => (v !== undefined ? v.toLocaleString() : "-"),
    sorter: (a, b) => (a.peOpenInterest || 0) - (b.peOpenInterest || 0),
    width: 80,
    align: "right",
  },
  {
    title: "IntraDay PCR",
    dataIndex: "intradayPCR",
    key: "intradayPCR",
    render: (_, record) => {
      const pePChange = record?.pePChangeInOI || 0;
      const cePChange = record?.cePChangeInOI || 0;
      
      if (cePChange === 0) return "-";
      
      const pcr = pePChange / cePChange;
      return isFinite(pcr) ? pcr.toFixed(2) : "-";
    },
    sorter: (a, b) => {
      const pePChangeA = a.pePChangeInOI || 0;
      const cePChangeA = a.cePChangeInOI || 0;
      const pePChangeB = b.pePChangeInOI || 0;
      const cePChangeB = b.cePChangeInOI || 0;
      
      const pcrA = cePChangeA === 0 ? 0 : pePChangeA / cePChangeA;
      const pcrB = cePChangeB === 0 ? 0 : pePChangeB / cePChangeB;
      
      const finalA = isFinite(pcrA) ? pcrA : 0;
      const finalB = isFinite(pcrB) ? pcrB : 0;
      
      return finalA - finalB;
    },
    width: 100,
    align: "right",
  },
  {
    title: "PCR",
    dataIndex: "pcr",
    key: "pcr",
    render: (_, record) => {
      const peOI = record?.peOpenInterest || 0;
      const ceOI = record?.ceOpenInterest || 0;
      
      if (ceOI === 0) return "-";
      
      const pcr = peOI / ceOI;
      return isFinite(pcr) ? pcr.toFixed(2) : "-";
    },
    sorter: (a, b) => {
      const peOIA = a.peOpenInterest || 0;
      const ceOIA = a.ceOpenInterest || 0;
      const peOIB = b.peOpenInterest || 0;
      const ceOIB = b.ceOpenInterest || 0;
      
      const pcrA = ceOIA === 0 ? 0 : peOIA / ceOIA;
      const pcrB = ceOIB === 0 ? 0 : peOIB / ceOIB;
      
      const finalA = isFinite(pcrA) ? pcrA : 0;
      const finalB = isFinite(pcrB) ? pcrB : 0;
      
      return finalA - finalB;
    },
    width: 80,
    align: "right",
  }
];

// Summary Table Columns
const summaryColumns = [
  {
    title: "Expiry Date",
    dataIndex: "expiryDate",
    key: "expiryDate",
    align: "center",
  },
  {
    title: "Call OI",
    dataIndex: "totalCEOI",
    key: "totalCEOI",
    render: (v) => v.toLocaleString(),
    align: "right",
  },
  {
    title: "Call CCOI",
    dataIndex: "totalCECCOI",
    key: "totalCECCOI",
    render: (v) => v.toLocaleString(),
    align: "right",
  },
  {
    title: "Call Volume",
    dataIndex: "totalCEVol",
    key: "totalCEVol",
    render: (v) => v.toLocaleString(),
    align: "right",
  },
  {
    title: "Put OI",
    dataIndex: "totalPEOI",
    key: "totalPEOI",
    render: (v) => v.toLocaleString(),
    align: "right",
  },
  {
    title: "Put CCOI",
    dataIndex: "totalPECCOI",
    key: "totalPECCOI",
    render: (v) => v.toLocaleString(),
    align: "right",
  },
  {
    title: "Put Volume",
    dataIndex: "totalPEVol",
    key: "totalPEVol",
    render: (v) => v.toLocaleString(),
    align: "right",
  },
  {
    title: "PCR (OI)",
    dataIndex: "pcrOI",
    key: "pcrOI",
    render: (v) => v.toFixed(2),
    align: "right",
  },
];

function getSummaryDataForExpiry(records, expiry) {
  const filtered = records.filter((rec) => rec.expiryDate === expiry);
  const summary = {
    expiryDate: expiry,
    totalCEOI: 0,
    totalCECCOI: 0,
    totalCEVol: 0,
    totalPEOI: 0,
    totalPECCOI: 0,
    totalPEVol: 0,
    pcrOI: 0,
  };

  for (const rec of filtered) {
    summary.totalCEOI += rec.ceOpenInterest || 0;
    summary.totalCECCOI += rec.ceChangeInOI || 0;
    summary.totalCEVol += rec.ceVolume || 0;

    summary.totalPEOI += rec.peOpenInterest || 0;
    summary.totalPECCOI += rec.peChangeInOI || 0;
    summary.totalPEVol += rec.peVolume || 0;
  }

  summary.pcrOI = summary.totalCEOI === 0 ? 0 : summary.totalPEOI / summary.totalCEOI;
  return summary;
}

function OptionChainTable() {
  const [loading, setLoading] = useState(true);
  const [rawRecords, setRawRecords] = useState([]);
  const [records, setRecords] = useState([]);
  const [meta, setMeta] = useState({ timestamp: "", underlyingValue: 0 });
  const [error, setError] = useState(null);

  const [minStrike, setMinStrike] = useState(null);
  const [maxStrike, setMaxStrike] = useState(null);
  const [expiryDates, setExpiryDates] = useState([]);
  const [selectedExpiry, setSelectedExpiry] = useState(undefined);

  const [summaryData, setSummaryData] = useState([]);
  const [historicalData, setHistoricalData] = useState([]);

  useEffect(() => {
    const source = new EventSource("http://localhost:4300/api/data");
    const timeout = setTimeout(() => {
      setError("No data received from SSE server.");
      setLoading(false);
      source.close();
    }, 5000);

    source.onmessage = (event) => {
      try {
        clearTimeout(timeout);
        const parsedData = JSON.parse(event.data);

        if (!Array.isArray(parsedData) || parsedData.length === 0) {
          return;
        }

        const recordToShow = parsedData[0];

        // Store historical data for charts
        setHistoricalData(prev => {
          const newData = [...prev, recordToShow];
          // Keep only last 100 data points to prevent memory issues
          return newData.slice(-100);
        });

        const map = {};
        const expirySet = new Set();

        recordToShow.data.forEach((item) => {
          const key = item.strikePrice + "-" + item.expiryDate;
          expirySet.add(item.expiryDate);
          if (!map[key]) {
            map[key] = {
              key,
              strikePrice: item.strikePrice,
              expiryDate: item.expiryDate,
              timestamp: recordToShow.timestamp,
              underlyingValue: recordToShow.underlyingValue,
            };
          }
          if (item.CE) {
            map[key].ceOpenInterest = item.CE.openInterest;
            map[key].ceChangeInOI = item.CE.changeinOpenInterest;
            map[key].cePChangeInOI = item.CE.pchangeinOpenInterest;
            map[key].ceVolume = item.CE.totalTradedVolume;
            map[key].ceIV = item.CE.impliedVolatility;
            map[key].ceLTP = item.CE.lastPrice;
          }
          if (item.PE) {
            map[key].peOpenInterest = item.PE.openInterest;
            map[key].peChangeInOI = item.PE.changeinOpenInterest;
            map[key].pePChangeInOI = item.PE.pchangeinOpenInterest;
            map[key].peVolume = item.PE.totalTradedVolume;
            map[key].peIV = item.PE.impliedVolatility;
            map[key].peLTP = item.PE.lastPrice;
          }
        });

        const allRecords = Object.values(map);
        const sortedExpiry = [...expirySet].sort();
        const summary = sortedExpiry.slice(0, 2).map((exp) => getSummaryDataForExpiry(allRecords, exp));

        setRawRecords(allRecords);
        setExpiryDates(sortedExpiry);
        setMeta({ timestamp: recordToShow.timestamp, underlyingValue: recordToShow.underlyingValue });
        setSummaryData(summary);
        setLoading(false);
      } catch (err) {
        console.error("Failed to parse SSE message", err);
        setError("Failed to parse SSE data.");
        setLoading(false);
      }
    };

    source.onerror = () => {
      setError("SSE connection error.");
      setLoading(false);
      source.close();
    };

    return () => {
      clearTimeout(timeout);
      source.close();
    };
  }, []);

  useEffect(() => {
    if (!selectedExpiry && expiryDates.length) {
      setSelectedExpiry(expiryDates[0]);
    }
  }, [expiryDates]);

  useEffect(() => {
    let filtered = rawRecords;
    if (selectedExpiry) {
      filtered = filtered.filter((rec) => rec.expiryDate === selectedExpiry);
    }
    if (minStrike !== null) {
      filtered = filtered.filter((rec) => rec.strikePrice >= minStrike);
    }
    if (maxStrike !== null) {
      filtered = filtered.filter((rec) => rec.strikePrice <= maxStrike);
    }
    setRecords(filtered);
  }, [rawRecords, minStrike, maxStrike, selectedExpiry]);

  return (
    <div className="w-screen h-screen p-4 overflow-auto bg-white">
      <Card
        className="mb-4 shadow-lg rounded-2xl"
        title={
          <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center">
            <span className="text-lg font-bold">NIFTY Option Chain</span>
            <span className="text-sm text-gray-500 mt-2 sm:mt-0">
              Last Updated: {meta.timestamp ? new Date(meta.timestamp).toLocaleString() : "--"}
            </span>
          </div>
        }
        extra={
          <span className="text-blue-700 font-semibold text-xl">
            Underlying Value: {meta.underlyingValue || "--"}
          </span>
        }
      >
        <Tabs defaultActiveKey="table" className="w-full">
          <TabPane tab="Option Chain Table" key="table">
            <div className="flex flex-col sm:flex-row flex-wrap items-center gap-4 mb-4">
              <div className="flex items-center gap-2">
                <span className="font-semibold text-gray-600">Strike Price:</span>
                <InputNumber min={0} placeholder="Min" size="small" value={minStrike} onChange={setMinStrike} className="!w-24" />
                <span>-</span>
                <InputNumber min={0} placeholder="Max" size="small" value={maxStrike} onChange={setMaxStrike} className="!w-24" />
                <Button size="small" onClick={() => {
                  setMinStrike(null);
                  setMaxStrike(null);
                  if (expiryDates.length) setSelectedExpiry(undefined);
                }}>
                  Reset
                </Button>
              </div>
              <div className="flex items-center gap-2">
                <span className="font-semibold text-gray-600">Expiry Date:</span>
                <Select
                  showSearch
                  size="small"
                  placeholder="Select expiry"
                  value={selectedExpiry}
                  onChange={(value) => setSelectedExpiry(value ?? undefined)}
                  className="!w-48"
                  optionFilterProp="children"
                  allowClear
                >
                  {expiryDates.map((exp) => (
                    <Option key={exp} value={exp}>{exp}</Option>
                  ))}
                </Select>
              </div>
            </div>

            {summaryData.length > 0 && (
              <>
                <Title level={5}>Expiry Summary</Title>
                <Table
                  pagination={false}
                  columns={summaryColumns}
                  dataSource={summaryData}
                  size="small"
                  bordered
                  rowKey="expiryDate"
                  className="mb-6"
                />
              </>
            )}

            {error && <Alert message={error} type="error" className="mb-4" />}
            {loading ? (
              <div className="flex justify-center items-center min-h-[200px]">
                <Spin size="large" />
              </div>
            ) : (
              <Table
                pagination={{ pageSize: 100 }}
                columns={columns}
                dataSource={records}
                size="small"
                scroll={{ x: "max-content" }}
                bordered
                rowClassName="text-xs text-center"
                className="rounded-2xl min-w-[1400px]"
                sticky
              />
            )}
          </TabPane>
          
          <TabPane tab="Line Charts" key="charts">
            <OILineChart data={historicalData} />
          </TabPane>
        </Tabs>
      </Card>
    </div>
  );
}

export default OptionChainTable;