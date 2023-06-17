# coding: utf-8
# Your code here!
#ビット全探索


def main():
    Items = []
    N, M = map(int, input().split())
    for _ in range(M):
        Infs = list(map(int, input().split()))[1:]
        Ns = [cnt for cnt in range(1, N+1)]
        Items.append(list(map(lambda cnt: 1 if cnt in Infs else 0, Ns)))

    Ans = list(map(int, input().split()))
    
    answear = 0
    for i in range(2**N):
        Comp = N*[0]
        for j in range(len(Comp)):
            if (i>>j) & 1:
                Comp[j] = 1
        # print(Comp)
        cnt = 0
        for Item, An in zip(Items, Ans):
            count = 0
            for item, comp in zip(Item, Comp):
                if item*comp == 1:
                    count += 1
            if count%2 == An:
                cnt += 1
        if cnt == M:
            answear += 1
    print(answear)
            
                
            
    


if __name__ == "__main__":
    main()
