# archeage-production-method

이 저장소는 아키에이지 제작법 사전을 YAML 형식으로 담고 있습니다.

YAML 형식으로 표현된 제작법은 다음과 같습니다.

```yaml
- name: 숙련공의 판금 투구
  material:
  - name: 수습공의 판금 투구
    quantity: 1
  - name: 달빛 아키움 조각
    quantity: 4
  - name: 철 주괴
    quantity: 12
  - name: 불투명한 연마제
    quantity: 2

- name: 마력이 봉인된 명인의 판금 투구
  material:
  - name: 숙련공의 판금 투구
    quantity: 1
  - name: 달빛 아키움 결정
    quantity: 6
  - name: 강도 높은 주괴
    quantity: 10
  - name: 거친 입자의 연마제
    quantity: 4
```

이 저장소는 Project-Arche Auction API에 검색 쿼리를 요청할 때 정확한 수량을 입력하기 위해 만들어졌습니다.