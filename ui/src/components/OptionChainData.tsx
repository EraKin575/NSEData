import { useEffect, useState } from "react";
import axios from "axios";
import { Table, Card, Spin, Alert, InputNumber, Select, Button } from "antd";
import { Color } from "antd/es/color-picker";

const { Option } = Select;

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
    render: (v: { toLocaleString: () => any; } | undefined) => (v !== undefined ? v.toLocaleString() : "-"),
    sorter: (a: { ceOpenInterest: any; }, b: { ceOpenInterest: any; }) => (a.ceOpenInterest || 0) - (b.ceOpenInterest || 0),
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
    sorter: (a: { peVolume: any; }, b: { peVolume: any; }) => (a.peVolume || 0) - (b.peVolume || 0),
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
      
      // Handle division by zero
      const pcrA = cePChangeA === 0 ? 0 : pePChangeA / cePChangeA;
      const pcrB = cePChangeB === 0 ? 0 : pePChangeB / cePChangeB;
      
      // Handle infinite values
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
      
      // Handle division by zero
      const pcrA = ceOIA === 0 ? 0 : peOIA / ceOIA;
      const pcrB = ceOIB === 0 ? 0 : peOIB / ceOIB;
      
      // Handle infinite values
      const finalA = isFinite(pcrA) ? pcrA : 0;
      const finalB = isFinite(pcrB) ? pcrB : 0;
      
      return finalA - finalB;
    },
    width: 80,
    align: "right",
  }
];

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
        const res = JSON.parse(event.data);
        const map = {};
        const expirySet = new Set();

        res.data.forEach((item) => {
          const key = item.strikePrice + "-" + item.expiryDate;
          expirySet.add(item.expiryDate);
          if (!map[key]) {
            map[key] = {
              key,
              strikePrice: item.strikePrice,
              expiryDate: item.expiryDate,
              timestamp: res.timestamp,
              underlyingValue: res.underlyingValue,
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
        setRawRecords(allRecords);
        setExpiryDates([...expirySet].sort());
        setMeta({
          timestamp: res.timestamp,
          underlyingValue: res.underlyingValue,
        });
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
        <div className="flex flex-col sm:flex-row
         flex-wrap items-center gap-4 mb-4">
          <div className="flex items-center gap-2">
            <span className="font-semibold text-gray-600">Strike Price:</span>
            <InputNumber
              min={0}
              placeholder="Min"
              size="small"
              value={minStrike}
              onChange={setMinStrike}
              className="!w-24"
            />
            <span>-</span>
            <InputNumber
              min={0}
              placeholder="Max"
              size="small"
              value={maxStrike}
              onChange={setMaxStrike}
              className="!w-24"
            />
            <Button
              size="small"
              onClick={() => {
                setMinStrike(null);
                setMaxStrike(null);
                if (expiryDates.length) setSelectedExpiry(undefined);
              }}
            >
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
                <Option key={exp} value={exp}>
                  {exp}
                </Option>
              ))}
            </Select>
          </div>
        </div>

        {error && <Alert message={error} type="error" className="mb-4" />}
        {loading ? (
          <div className="flex justify-center items-center min-h-[200px]">
            <Spin size="large" />
          </div>
        ) : (
          <Table
            pagination={{
              pageSize: 100
            }}
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
      </Card>
    </div>
  );
}

export default OptionChainTable;
