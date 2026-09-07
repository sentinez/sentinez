import { ResourceView } from './components/resource-view';
export default async function ResourcePage({ params }) {
    const { domain } = await params;
    return <ResourceView domain={domain}/>;
}
