import { useEffect, useState } from "react";
import { Table, Card, Spin, Alert, InputNumber, Select, Button, Tabs, Typography } from "antd";
import OILineChart from "./charts"; // Assuming this file exists

const { Option } = Select;
const { Title } = Typography;
const { TabPane } = Tabs;

// Column definitions remain the same
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
    defaultSortOrder: 'descend',
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
    render: (v) => (v ? v.toFixed(2) : "-"),
    sorter: (a, b) => (a.underlyingValue || 0) - (b.underlyingValue || 0),
    width: 80,
    align: "center",
  },
  {
    title: <span>Strike Price</span>,
    dataIndex: "strikePrice",
    key: "strikePrice",
    render: (v) => (
      <div className="bg-amber-200 p-0">
        <strong className="text-amber-600 p-0.5">{v?.toLocaleString()}</strong>
      </div>
    ),
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
      const pcrA = (a.cePChangeInOI || 0) === 0 ? 0 : (a.pePChangeInOI || 0) / a.cePChangeInOI;
      const pcrB = (b.cePChangeInOI || 0) === 0 ? 0 : (b.pePChangeInOI || 0) / b.cePChangeInOI;
      return (isFinite(pcrA) ? pcrA : 0) - (isFinite(pcrB) ? pcrB : 0);
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
      const pcrA = (a.ceOpenInterest || 0) === 0 ? 0 : (a.peOpenInterest || 0) / a.ceOpenInterest;
      const pcrB = (b.ceOpenInterest || 0) === 0 ? 0 : (b.peOpenInterest || 0) / b.ceOpenInterest;
      return (isFinite(pcrA) ? pcrA : 0) - (isFinite(pcrB) ? pcrB : 0);
    },
    width: 80,
    align: "right",
  },
];

const summaryColumns = [
  { title: "Expiry Date", dataIndex: "expiryDate", key: "expiryDate", align: "center" },
  { title: "Call OI", dataIndex: "totalCEOI", key: "totalCEOI", render: (v) => v.toLocaleString(), align: "right" },
  { title: "Call CCOI", dataIndex: "totalCECCOI", key: "totalCECCOI", render: (v) => v.toLocaleString(), align: "right" },
  { title: "Call Volume", dataIndex: "totalCEVol", key: "totalCEVol", render: (v) => v.toLocaleString(), align: "right" },
  { title: "Put OI", dataIndex: "totalPEOI", key: "totalPEOI", render: (v) => v.toLocaleString(), align: "right" },
  { title: "Put CCOI", dataIndex: "totalPECCOI", key: "totalPECCOI", render: (v) => v.toLocaleString(), align: "right" },
  { title: "Put Volume", dataIndex: "totalPEVol", key: "totalPEVol", render: (v) => v.toLocaleString(), align: "right" },
  { title: "PCR (OI)", dataIndex: "pcrOI", key: "pcrOI", render: (v) => v.toFixed(2), align: "right" },
];

function getSummaryDataForExpiry(records, expiry) {
  const filtered = records.filter((rec) => rec.expiryDate === expiry);
  const summary = {
    expiryDate: expiry,
    totalCEOI: 0, totalCECCOI: 0, totalCEVol: 0,
    totalPEOI: 0, totalPECCOI: 0, totalPEVol: 0,
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
  const [rawRecords, setRawRecords] = useState([]);
  const [expiryDates, setExpiryDates] = useState([]);
  const [summaryData, setSummaryData] = useState([]);
  const [meta, setMeta] = useState({ timestamp: "", underlyingValue: 0 });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [records, setRecords] = useState([]);
  const [minStrike, setMinStrike] = useState(null);
  const [maxStrike, setMaxStrike] = useState(null);
  const [selectedExpiry, setSelectedExpiry] = useState(undefined);
  const [historicalData, setHistoricalData] = useState([]);
  const [isDataReceived, setDataReceived] = useState(false);
  const [connectionStatus, setConnectionStatus] = useState('connecting');

  useEffect(() => {
    let eventSource = null;
    let retryTimeoutId = null;
    let retryCount = 0;
    const maxRetries = 5;

    const connectSSE = () => {
      setConnectionStatus('connecting');
      setError(null);
      
      try {
        eventSource = new EventSource("http://localhost:4300/api/data");
        
        eventSource.onopen = () => {
          console.log('SSE connection opened');
          setConnectionStatus('connected');
          retryCount = 0;
        };

        eventSource.onmessage = (event) => {
          try {
            setLoading(false);
            setConnectionStatus('receiving');

            if (!event.data || event.data.trim() === '') {
              console.warn('Received empty SSE data');
              return;
            }

            const parsedData = JSON.parse(event.data);
            if (!parsedData) {
              console.warn('Parsed data is null or undefined');
              return;
            }
            
            let dataArray, latestTimestamp, latestUnderlyingValue;
            
            if (Array.isArray(parsedData) && parsedData.length > 0) {
              const allRecordsData = [];
              const timestampsSet = new Set();
              
              for (const record of parsedData) {
                if (record && record.data && Array.isArray(record.data)) {
                  const recordTimestamp = record.timestamp || record.TimeStamp;
                  const recordUnderlyingValue = record.underlyingValue || record.UnderlyingValue;
                  timestampsSet.add(recordTimestamp);
                  
                  const recordDataWithMeta = record.data.map(item => ({
                    ...item,
                    recordTimestamp: recordTimestamp,
                    recordUnderlyingValue: recordUnderlyingValue
                  }));
                  allRecordsData.push(...recordDataWithMeta);
                }
              }
              
              dataArray = allRecordsData;
              const sortedTimestamps = Array.from(timestampsSet).sort((a, b) => new Date(b).getTime() - new Date(a).getTime());
              latestTimestamp = sortedTimestamps[0];
              const latestRecord = parsedData.find(r => (r.timestamp || r.TimeStamp) === latestTimestamp);
              latestUnderlyingValue = latestRecord?.underlyingValue || latestRecord?.UnderlyingValue;
            } else {
              console.error('Unexpected data structure:', parsedData);
              setError('Unexpected data structure from server.');
              return;
            }

            if (!dataArray || dataArray.length === 0) {
              console.warn('No data array found or empty array');
              return;
            }

            setDataReceived(true);
            const recordMap = {};
            const expirySet = new Set();

            // =================================================================
            // START OF FIX: This loop now correctly processes all historical data.
            // =================================================================
            for (const item of dataArray) {
              if (!item || typeof item !== 'object') continue;

              const strikePrice = item.strikePrice;
              const expiryDate = item.expiryDate;
              
              // FIX #1: Get the timestamp that belongs to THIS specific item.
              // This was the source of the bug, as it was using a generic timestamp before.
              const itemTimestamp = item.recordTimestamp;
              
              if (!strikePrice || !expiryDate || !itemTimestamp) {
                console.warn('Missing required fields in item:', item);
                continue;
              }

              // FIX #2: Create a key that is unique for every single data point
              // by including the timestamp. This prevents overwriting historical data.
              const key = `${strikePrice}-${expiryDate}-${itemTimestamp}`;
              expirySet.add(expiryDate);

              // This check now correctly identifies unique data points.
              if (!recordMap[key]) {
                recordMap[key] = {
                  key,
                  strikePrice: Number(strikePrice),
                  expiryDate: expiryDate,
                  // FIX #3: Assign the item's own timestamp and underlying value.
                  timestamp: itemTimestamp,
                  underlyingValue: Number(item.recordUnderlyingValue) || null,
                };
              }

              // Assign CE (Call) data
              if (item.CE && typeof item.CE === 'object') {
                Object.assign(recordMap[key], {
                  ceOpenInterest: Number(item.CE.openInterest || 0),
                  ceChangeInOI: Number(item.CE.changeinOpenInterest || 0),
                  cePChangeInOI: Number(item.CE.pchangeinOpenInterest || 0),
                  ceVolume: Number(item.CE.totalTradedVolume || 0),
                  ceIV: Number(item.CE.impliedVolatility || 0),
                  ceLTP: Number(item.CE.lastPrice || 0),
                });
              }

              // Assign PE (Put) data
              if (item.PE && typeof item.PE === 'object') {
                Object.assign(recordMap[key], {
                  peOpenInterest: Number(item.PE.openInterest || 0),
                  peChangeInOI: Number(item.PE.changeinOpenInterest || 0),
                  pePChangeInOI: Number(item.PE.pchangeinOpenInterest || 0),
                  peVolume: Number(item.PE.totalTradedVolume || 0),
                  peIV: Number(item.PE.impliedVolatility || 0),
                  peLTP: Number(item.PE.lastPrice || 0),
                });
              }
            }
            // =================================================================
            // END OF FIX
            // =================================================================

            const allRecords = Object.values(recordMap);
            const sortedExpiry = [...expirySet].sort();

            if (allRecords.length === 0) {
              setError('No valid records were processed from the data');
              return;
            }

            setRawRecords(allRecords);
            setExpiryDates(sortedExpiry);
            setMeta({
              timestamp: latestTimestamp,
              underlyingValue: Number(latestUnderlyingValue) || 0,
            });
            setError(null);
            setConnectionStatus('connected');

          } catch (err) {
            console.error("SSE MESSAGE PROCESSING ERROR:", err);
            setError(`Data processing error: ${err.message}`);
            setConnectionStatus('error');
          }
        };

        eventSource.onerror = () => {
          console.error('SSE connection error');
          setConnectionStatus('error');
          if (eventSource.readyState === EventSource.CLOSED) {
            setError("Connection closed. Retrying...");
            retryConnection();
          }
          setLoading(false);
        };

      } catch (err) {
        console.error('Error creating EventSource:', err);
        setError(`Connection error: ${err.message}`);
        setLoading(false);
        setConnectionStatus('error');
      }
    };

    const retryConnection = () => {
      if (retryCount < maxRetries) {
        retryCount++;
        const delay = Math.min(1000 * Math.pow(2, retryCount), 30000);
        console.log(`Retrying connection in ${delay / 1000}s...`);
        retryTimeoutId = setTimeout(connectSSE, delay);
      } else {
        setError("Failed to connect after multiple attempts.");
      }
    };

    connectSSE();

    return () => {
      if (eventSource) eventSource.close();
      if (retryTimeoutId) clearTimeout(retryTimeoutId);
    };
  }, []);

  useEffect(() => {
    if (!selectedExpiry && expiryDates.length) {
      setSelectedExpiry(expiryDates[0]);
    }
  }, [expiryDates, selectedExpiry]);

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

  useEffect(() => {
    if (rawRecords.length > 0 && expiryDates.length > 0) {
      const summary = expiryDates
        .slice(0, 2)
        .map((exp) => getSummaryDataForExpiry(rawRecords, exp));
      setSummaryData(summary);
    }
  }, [rawRecords, expiryDates]);

  const statusInfo = {
    connecting: { message: "Connecting to data stream...", type: "info" },
    connected: { message: isDataReceived ? "Connected, streaming data" : "Connected, waiting for data...", type: "success" },
    receiving: { message: "Live data stream active", type: "success" },
    error: { message: error || "Connection error", type: "error" },
  }[connectionStatus] || { message: "Unknown status", type: "warning" };

  return (
    <div className="w-screen h-screen p-4 overflow-auto bg-white">
      <Card
        className="mb-4 shadow-lg rounded-2xl"
        title={
          <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center">
            <span className="text-lg font-bold">NIFTY Option Chain</span>
            <span className="text-sm text-gray-500 mt-2 sm:mt-0">
              Last Updated:{" "}
              {meta.timestamp ? new Date(meta.timestamp).toLocaleString() : "--"}
            </span>
          </div>
        }
        extra={
          <div className="flex flex-col items-end gap-1">
            <span className="text-blue-700 font-semibold text-xl">
              Underlying Value: {meta.underlyingValue || "--"}
            </span>
            <span className={`text-xs px-2 py-1 rounded ${
              statusInfo.type === 'success' ? 'bg-green-100 text-green-800' :
              statusInfo.type === 'info' ? 'bg-blue-100 text-blue-800' :
              'bg-red-100 text-red-800'
            }`}>
              {connectionStatus.toUpperCase()}
            </span>
          </div>
        }
      >
        <Tabs defaultActiveKey="table" className="w-full">
          <TabPane tab="Option Chain Table" key="table">
            <div className="flex flex-col sm:flex-row flex-wrap items-center gap-4 mb-4">
              <div className="flex items-center gap-2">
                <span className="font-semibold text-gray-600">Strike Price:</span>
                <InputNumber size="small" placeholder="Min" value={minStrike} onChange={setMinStrike} className="!w-24" />
                <span>-</span>
                <InputNumber size="small" placeholder="Max" value={maxStrike} onChange={setMaxStrike} className="!w-24" />
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
                  allowClear
                >
                  {expiryDates.map((exp) => <Option key={exp} value={exp}>{exp}</Option>)}
                </Select>
              </div>
              <Button size="small" onClick={() => { setMinStrike(null); setMaxStrike(null); setSelectedExpiry(undefined); }}>
                Reset Filters
              </Button>
            </div>

            {summaryData.length > 0 && (
              <>
                <Title level={5}>Expiry Summary</Title>
                <Table pagination={false} columns={summaryColumns} dataSource={summaryData} size="small" bordered rowKey="expiryDate" className="mb-6" />
              </>
            )}

            <Alert message={statusInfo.message} type={statusInfo.type} className="mb-4" />
            
            {loading ? (
              <div className="flex justify-center items-center min-h-[200px]"><Spin size="large" /></div>
            ) : !isDataReceived ? (
              <div className="flex justify-center items-center min-h-[200px]">
                <Alert message="No data received yet" description="Please ensure the server is running and streaming data." type="warning" showIcon />
              </div>
            ) : (
              <Table
                pagination={{ pageSize: 100, showSizeChanger: true, pageSizeOptions: ['50', '100', '200', '500'] }}
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
