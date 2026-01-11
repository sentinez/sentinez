type Props = {
  params: {
    id: string;
  };
};

export default function ResourceSetting({ params }: Props) {
  return <div>User ID: {params.id}</div>;
}
