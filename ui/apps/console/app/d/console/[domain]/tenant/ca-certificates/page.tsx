import View from './view';
import { pageDelay } from '@/lib/delays';

export default async function Page() {
  await pageDelay();
  return <View />;
}
