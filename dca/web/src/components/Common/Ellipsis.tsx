import { Typography } from 'antd';

const { Text } = Typography;

export const EllipsisMiddle = ({ suffixCount, children, maxWidth }) => {
  const start = children.slice(0, children.length - suffixCount).trim();
  const suffix = children.slice(-suffixCount).trim();
  return (
    <Text style={{ maxWidth }} ellipsis={{ suffix }}>
      {start}
    </Text>
  );
};
